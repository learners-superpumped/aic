// Package api is the single entry point for all backend HTTP calls.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// doOnce performs a single HTTP request and returns its status, response
// headers, body, and any transport-level error.
func (c *Client) doOnce(ctx context.Context, method, path string, body any) (int, http.Header, []byte, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
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

// doRequest sends a request with one transparent 401 refresh-and-retry, shared
// by do (JSON) and doRaw (binary).
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (int, http.Header, []byte, error) {
	status, hdr, data, err := c.doOnce(ctx, method, path, body)
	if err != nil {
		return 0, nil, nil, err
	}
	if status == http.StatusUnauthorized && c.refreshFn != nil {
		if rerr := c.refresh(ctx); rerr == nil {
			status, hdr, data, err = c.doOnce(ctx, method, path, body)
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
		const hint = "your session has expired — run `aic login`"
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
	status, _, data, err := c.doRequest(ctx, method, path, body)
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
	status, hdr, data, err := c.doRequest(ctx, method, path, nil)
	if err != nil {
		return nil, nil, err
	}
	if status >= 400 {
		return nil, nil, apiError(status, data)
	}
	return data, hdr, nil
}

// doOnceBytes performs a single request with a raw byte body and an explicit
// Content-Type, returning the response status, headers, and body.
func (c *Client) doOnceBytes(ctx context.Context, method, path string, body []byte, contentType string) (int, http.Header, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
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

// doBytes sends a raw byte body with one transparent 401 refresh-and-retry — the
// binary-upload sibling of do.
func (c *Client) doBytes(ctx context.Context, method, path string, body []byte, contentType string, out any) error {
	status, _, data, err := c.doOnceBytes(ctx, method, path, body, contentType)
	if err != nil {
		return err
	}
	if status == http.StatusUnauthorized && c.refreshFn != nil {
		if rerr := c.refresh(ctx); rerr == nil {
			status, _, data, err = c.doOnceBytes(ctx, method, path, body, contentType)
			if err != nil {
				return err
			}
		}
	}
	if status >= 400 {
		return apiError(status, data)
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}
