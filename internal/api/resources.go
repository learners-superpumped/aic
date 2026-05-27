package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// --- Billing (scoped to a team) ---

func teamBillingPath(teamID string) string {
	return "/v1/teams/" + url.PathEscape(teamID) + "/billing"
}

func (c *Client) StartCardSession(ctx context.Context, teamID string) (*Session, error) {
	var s Session
	return &s, c.do(ctx, http.MethodPost, teamBillingPath(teamID)+"/card-sessions", nil, &s)
}

func (c *Client) PollCardSession(ctx context.Context, teamID, id string) (*Session, error) {
	var s Session
	return &s, c.do(ctx, http.MethodGet, teamBillingPath(teamID)+"/card-sessions/"+url.PathEscape(id), nil, &s)
}

func (c *Client) ListCards(ctx context.Context, teamID string) ([]Card, error) {
	var out []Card
	return out, c.do(ctx, http.MethodGet, teamBillingPath(teamID)+"/cards", nil, &out)
}

func (c *Client) Balance(ctx context.Context, teamID string) (*CreditBalance, error) {
	var b CreditBalance
	return &b, c.do(ctx, http.MethodGet, teamBillingPath(teamID)+"/balance", nil, &b)
}

func (c *Client) History(ctx context.Context, teamID string) ([]LedgerEntry, error) {
	var out []LedgerEntry
	return out, c.do(ctx, http.MethodGet, teamBillingPath(teamID)+"/history", nil, &out)
}

func (c *Client) Topup(ctx context.Context, teamID string, amountCents int64) (*TopupResult, error) {
	var res TopupResult
	return &res, c.do(ctx, http.MethodPost, teamBillingPath(teamID)+"/topup",
		map[string]int64{"amount_cents": amountCents}, &res)
}

// --- Teams ---

func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	var out []Team
	return out, c.do(ctx, http.MethodGet, "/v1/teams", nil, &out)
}

func (c *Client) CreateTeam(ctx context.Context, name string) (*Team, error) {
	var t Team
	return &t, c.do(ctx, http.MethodPost, "/v1/teams",
		map[string]string{"name": name}, &t)
}

func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	var t Team
	return &t, c.do(ctx, http.MethodGet, "/v1/teams/"+url.PathEscape(id), nil, &t)
}

// --- Projects (scoped to a team) ---

func teamProjectsPath(teamID string) string {
	return "/v1/teams/" + url.PathEscape(teamID) + "/projects"
}

func (c *Client) ListProjects(ctx context.Context, teamID string) ([]Project, error) {
	var out []Project
	return out, c.do(ctx, http.MethodGet, teamProjectsPath(teamID), nil, &out)
}

func (c *Client) CreateProject(ctx context.Context, teamID, name string) (*Project, error) {
	var p Project
	return &p, c.do(ctx, http.MethodPost, teamProjectsPath(teamID),
		map[string]string{"name": name}, &p)
}

func (c *Client) GetProject(ctx context.Context, teamID, id string) (*Project, error) {
	var p Project
	return &p, c.do(ctx, http.MethodGet, teamProjectsPath(teamID)+"/"+url.PathEscape(id), nil, &p)
}

func (c *Client) DeleteProject(ctx context.Context, teamID, id string) error {
	return c.do(ctx, http.MethodDelete, teamProjectsPath(teamID)+"/"+url.PathEscape(id), nil, nil)
}

// --- Domains (scoped to a team's project) ---

func teamDomainsPath(teamID, projectID string) string {
	return "/v1/teams/" + url.PathEscape(teamID) + "/projects/" + url.PathEscape(projectID) + "/domains"
}

func (c *Client) SearchDomains(ctx context.Context, teamID, projectID, query string) ([]DomainSearchResult, error) {
	var out []DomainSearchResult
	return out, c.do(ctx, http.MethodGet, teamDomainsPath(teamID, projectID)+"/search?q="+url.QueryEscape(query), nil, &out)
}

func (c *Client) BuyDomain(ctx context.Context, teamID, projectID, domain string, years int, autoRenew bool) (*Domain, error) {
	var d Domain
	return &d, c.do(ctx, http.MethodPost, teamDomainsPath(teamID, projectID),
		map[string]any{"name": domain, "years": years, "auto_renew": autoRenew}, &d)
}

func (c *Client) RenewDomain(ctx context.Context, teamID, projectID, domain string, years int) (*Domain, error) {
	var d Domain
	return &d, c.do(ctx, http.MethodPost, teamDomainsPath(teamID, projectID)+"/"+url.PathEscape(domain)+"/renew",
		map[string]any{"years": years}, &d)
}

func (c *Client) ListDomains(ctx context.Context, teamID, projectID string) ([]Domain, error) {
	var out []Domain
	return out, c.do(ctx, http.MethodGet, teamDomainsPath(teamID, projectID), nil, &out)
}

func (c *Client) GetDomain(ctx context.Context, teamID, projectID, domain string) (*Domain, error) {
	var d Domain
	return &d, c.do(ctx, http.MethodGet, teamDomainsPath(teamID, projectID)+"/"+url.PathEscape(domain), nil, &d)
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
