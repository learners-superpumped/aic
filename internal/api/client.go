// Package api is the single entry point for all backend HTTP calls.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Error is a structured backend error.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("request failed with status %d", e.Status)
}

// Client calls the backend API with a bearer token. It can transparently
// refresh an expired access token on a 401 response when configured via
// WithRefresh.
type Client struct {
	baseURL   string
	token     string
	refreshFn func(context.Context) (*Tokens, error)
	onRefresh func(*Tokens)
	http      *http.Client
}

// New returns a Client for baseURL authenticating with token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// WithRefresh enables transparent access-token refresh on 401. refreshFn obtains
// new tokens (e.g. via the OIDC token endpoint); onRefresh (if non-nil) persists
// them. Returns c for chaining.
func (c *Client) WithRefresh(refreshFn func(context.Context) (*Tokens, error), onRefresh func(*Tokens)) *Client {
	c.refreshFn = refreshFn
	c.onRefresh = onRefresh
	return c
}

// listPath builds a list endpoint path with the standard limit/cursor query and
// any extra params, encoding them in one place so every List* method shares the
// same pagination wire format. Empty extra values are skipped.
func listPath(base string, limit int, cursor string, extra url.Values) string {
	q := url.Values{}
	for k, vs := range extra {
		for _, v := range vs {
			if v != "" {
				q.Add(k, v)
			}
		}
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if e := q.Encode(); e != "" {
		return base + "?" + e
	}
	return base
}

// bodyFn produces a fresh request body reader and its Content-Type for each
// attempt (nil reader, "" type → no body). A factory is required because a
// retried request needs a new reader.
type bodyFn func() (io.Reader, string, error)

// jsonBody returns a bodyFn that JSON-encodes body (nil → no body).
func jsonBody(body any) bodyFn {
	return func() (io.Reader, string, error) {
		if body == nil {
			return nil, "", nil
		}
		b, err := json.Marshal(body)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(b), "application/json", nil
	}
}

// rawBody returns a bodyFn for a raw byte body with an explicit Content-Type.
func rawBody(body []byte, contentType string) bodyFn {
	return func() (io.Reader, string, error) {
		return bytes.NewReader(body), contentType, nil
	}
}

// doOnce performs a single HTTP request and returns its status, response
// headers, body, and any transport-level error. extraHeaders, when non-nil, are
// set on the request (e.g. a per-request Idempotency-Key).
func (c *Client) doOnce(ctx context.Context, method, path string, mk bodyFn, extraHeaders http.Header) (int, http.Header, []byte, error) {
	reader, contentType, err := mk()
	if err != nil {
		return 0, nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, vs := range extraHeaders {
		for _, v := range vs {
			req.Header.Set(k, v)
		}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, err
	}
	return resp.StatusCode, resp.Header, data, nil
}

// doRequest sends a request with one transparent 401 refresh-and-retry — the
// single retry path shared by do (JSON), doRaw (binary GET), and doBytes
// (binary upload).
func (c *Client) doRequest(ctx context.Context, method, path string, mk bodyFn) (int, http.Header, []byte, error) {
	return c.doRequestWithHeaders(ctx, method, path, mk, nil)
}

// doRequestWithHeaders is doRequest with per-request headers; the same key is
// reused on the 401 refresh-and-retry, so a client-supplied Idempotency-Key
// survives the transparent retry.
func (c *Client) doRequestWithHeaders(ctx context.Context, method, path string, mk bodyFn, extraHeaders http.Header) (int, http.Header, []byte, error) {
	status, hdr, data, err := c.doOnce(ctx, method, path, mk, extraHeaders)
	if err != nil {
		return 0, nil, nil, err
	}
	if status == http.StatusUnauthorized && c.refreshFn != nil {
		if rerr := c.refresh(ctx); rerr == nil {
			status, hdr, data, err = c.doOnce(ctx, method, path, mk, extraHeaders)
			if err != nil {
				return 0, nil, nil, err
			}
		}
	}
	return status, hdr, data, nil
}

// apiError builds an *Error from a >=400 response body, adding a login hint on 401.
func apiError(status int, data []byte) error {
	apiErr := &Error{Status: status}
	_ = json.Unmarshal(data, apiErr)
	if status == http.StatusUnauthorized {
		const hint = "your session has expired — run `aic auth login`"
		if apiErr.Message == "" {
			apiErr.Message = hint
		} else {
			apiErr.Message = apiErr.Message + " (" + hint + ")"
		}
	}
	return apiErr
}

func (c *Client) refresh(ctx context.Context) error {
	t, err := c.refreshFn(ctx)
	if err != nil {
		return err
	}
	c.token = t.AccessToken
	if c.onRefresh != nil {
		c.onRefresh(t)
	}
	return nil
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	return c.doWithHeaders(ctx, method, path, body, out, nil)
}

// doWithHeaders is do with per-request headers (e.g. Idempotency-Key).
func (c *Client) doWithHeaders(ctx context.Context, method, path string, body, out any, extraHeaders http.Header) error {
	status, _, data, err := c.doRequestWithHeaders(ctx, method, path, jsonBody(body), extraHeaders)
	if err != nil {
		return err
	}
	if status >= 400 {
		return apiError(status, data)
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

// doRaw performs a GET and returns the raw response bytes and headers without
// JSON-decoding — for binary endpoints like attachment download.
func (c *Client) doRaw(ctx context.Context, method, path string) ([]byte, http.Header, error) {
	status, hdr, data, err := c.doRequest(ctx, method, path, jsonBody(nil))
	if err != nil {
		return nil, nil, err
	}
	if status >= 400 {
		return nil, nil, apiError(status, data)
	}
	return data, hdr, nil
}

// doBytes sends a raw byte body with one transparent 401 refresh-and-retry — the
// binary-upload sibling of do.
func (c *Client) doBytes(ctx context.Context, method, path string, body []byte, contentType string, out any) error {
	status, _, data, err := c.doRequest(ctx, method, path, rawBody(body, contentType))
	if err != nil {
		return err
	}
	if status >= 400 {
		return apiError(status, data)
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}
