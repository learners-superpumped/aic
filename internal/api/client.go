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

// doOnce performs a single HTTP request and returns its status, body, and any
// transport-level error.
func (c *Client) doOnce(ctx context.Context, method, path string, body any) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, data, nil
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
	status, data, err := c.doOnce(ctx, method, path, body)
	if err != nil {
		return err
	}
	// On a 401, attempt a one-time transparent refresh and retry.
	if status == http.StatusUnauthorized && c.refreshFn != nil {
		if rerr := c.refresh(ctx); rerr == nil {
			status, data, err = c.doOnce(ctx, method, path, body)
			if err != nil {
				return err
			}
		}
	}
	if status >= 400 {
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
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}
