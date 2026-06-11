package api

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// APIKey is an AIC API key as returned by the backend. Key carries the raw
// secret only in the create response and is never returned by list.
type APIKey struct {
	ID         string     `json:"id"`
	Prefix     string     `json:"prefix"`
	Scopes     []string   `json:"scopes"`
	Name       string     `json:"name,omitempty"`
	ProjectID  string     `json:"project_id,omitempty"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	Status     string     `json:"status"`
	Key        string     `json:"key,omitempty"`
}

type CreateAPIKeyRequest struct {
	Scopes     []string `json:"scopes,omitempty"`
	FullAccess bool     `json:"full_access,omitempty"`
	ProjectID  string   `json:"project_id,omitempty"`
	Name       string   `json:"name,omitempty"`
	ExpiresIn  int64    `json:"expires_in_seconds,omitempty"`
}

func (c *Client) CreateAPIKey(ctx context.Context, teamID string, req CreateAPIKeyRequest) (APIKey, error) {
	var k APIKey
	err := c.do(ctx, http.MethodPost, "/v1/teams/"+url.PathEscape(teamID)+"/keys", req, &k)
	return k, err
}

func (c *Client) ListAPIKeys(ctx context.Context, teamID, projectID string, limit int, cursor string) (Page[APIKey], error) {
	base := "/v1/teams/" + url.PathEscape(teamID) + "/keys"
	path := listPath(base, limit, cursor, url.Values{"project": {projectID}})
	var page Page[APIKey]
	err := c.do(ctx, http.MethodGet, path, nil, &page)
	return page, err
}

func (c *Client) RevokeAPIKey(ctx context.Context, teamID, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/teams/"+url.PathEscape(teamID)+"/keys/"+url.PathEscape(id), nil, nil)
}
