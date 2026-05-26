package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// --- Billing ---

func (c *Client) StartCardSession(ctx context.Context) (*Session, error) {
	var s Session
	return &s, c.do(ctx, http.MethodPost, "/v1/billing/card-sessions", nil, &s)
}

func (c *Client) PollCardSession(ctx context.Context, id string) (*Session, error) {
	var s Session
	return &s, c.do(ctx, http.MethodGet, "/v1/billing/card-sessions/"+url.PathEscape(id), nil, &s)
}

func (c *Client) ListCards(ctx context.Context) ([]Card, error) {
	var out []Card
	return out, c.do(ctx, http.MethodGet, "/v1/billing/cards", nil, &out)
}

func (c *Client) BillingStatus(ctx context.Context) (*BillingStatus, error) {
	var s BillingStatus
	return &s, c.do(ctx, http.MethodGet, "/v1/billing/status", nil, &s)
}

// --- Projects ---

func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var out []Project
	return out, c.do(ctx, http.MethodGet, "/v1/projects", nil, &out)
}

func (c *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	var p Project
	return &p, c.do(ctx, http.MethodPost, "/v1/projects",
		map[string]string{"name": name}, &p)
}

func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	var p Project
	return &p, c.do(ctx, http.MethodGet, "/v1/projects/"+url.PathEscape(id), nil, &p)
}

func (c *Client) DeleteProject(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/projects/"+url.PathEscape(id), nil, nil)
}

// --- Domains ---

func (c *Client) SearchDomains(ctx context.Context, pid, query string) ([]DomainSearchResult, error) {
	var out []DomainSearchResult
	path := fmt.Sprintf("/v1/projects/%s/domains/search?q=%s", url.PathEscape(pid), url.QueryEscape(query))
	return out, c.do(ctx, http.MethodGet, path, nil, &out)
}

func (c *Client) BuyDomain(ctx context.Context, pid, domain string) (*Domain, error) {
	var d Domain
	return &d, c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/projects/%s/domains", url.PathEscape(pid)),
		map[string]string{"domain": domain}, &d)
}

func (c *Client) ListDomains(ctx context.Context, pid string) ([]Domain, error) {
	var out []Domain
	return out, c.do(ctx, http.MethodGet, fmt.Sprintf("/v1/projects/%s/domains", url.PathEscape(pid)), nil, &out)
}

func (c *Client) GetDomain(ctx context.Context, pid, domain string) (*Domain, error) {
	var d Domain
	path := fmt.Sprintf("/v1/projects/%s/domains/%s", url.PathEscape(pid), url.PathEscape(domain))
	return &d, c.do(ctx, http.MethodGet, path, nil, &d)
}

// --- Inboxes ---

func (c *Client) CreateInbox(ctx context.Context, pid, address string) (*Inbox, error) {
	var in Inbox
	return &in, c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/projects/%s/inboxes", url.PathEscape(pid)),
		map[string]string{"address": address}, &in)
}

func (c *Client) ListInboxes(ctx context.Context, pid string) ([]Inbox, error) {
	var out []Inbox
	return out, c.do(ctx, http.MethodGet, fmt.Sprintf("/v1/projects/%s/inboxes", url.PathEscape(pid)), nil, &out)
}

func (c *Client) GetInbox(ctx context.Context, pid, address string) (*Inbox, error) {
	var in Inbox
	path := fmt.Sprintf("/v1/projects/%s/inboxes/%s", url.PathEscape(pid), url.PathEscape(address))
	return &in, c.do(ctx, http.MethodGet, path, nil, &in)
}

func (c *Client) DeleteInbox(ctx context.Context, pid, address string) error {
	path := fmt.Sprintf("/v1/projects/%s/inboxes/%s", url.PathEscape(pid), url.PathEscape(address))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// --- Messages ---

func (c *Client) SendMessage(ctx context.Context, pid, address, to, subject, body string) (*Message, error) {
	var m Message
	path := fmt.Sprintf("/v1/projects/%s/inboxes/%s/messages", url.PathEscape(pid), url.PathEscape(address))
	return &m, c.do(ctx, http.MethodPost, path,
		map[string]string{"to": to, "subject": subject, "body": body}, &m)
}

func (c *Client) ListMessages(ctx context.Context, pid, address string) ([]Message, error) {
	var out []Message
	path := fmt.Sprintf("/v1/projects/%s/inboxes/%s/messages", url.PathEscape(pid), url.PathEscape(address))
	return out, c.do(ctx, http.MethodGet, path, nil, &out)
}

func (c *Client) GetMessage(ctx context.Context, pid, address, id string) (*Message, error) {
	var m Message
	path := fmt.Sprintf("/v1/projects/%s/inboxes/%s/messages/%s",
		url.PathEscape(pid), url.PathEscape(address), url.PathEscape(id))
	return &m, c.do(ctx, http.MethodGet, path, nil, &m)
}
