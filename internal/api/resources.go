package api

import (
	"context"
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

func (c *Client) BuyDomain(ctx context.Context, teamID, projectID, domain string, years int, autoRenew bool, contactName string) (*Domain, error) {
	var d Domain
	body := map[string]any{"name": domain, "years": years, "auto_renew": autoRenew}
	if contactName != "" {
		body["contact_name"] = contactName
	}
	return &d, c.do(ctx, http.MethodPost, teamDomainsPath(teamID, projectID), body, &d)
}

// --- Domain Contacts (per-team profiles for WHOIS) ---

func teamDomainContactsPath(teamID string) string {
	return "/v1/teams/" + url.PathEscape(teamID) + "/domain-contacts"
}

func (c *Client) CreateDomainContact(ctx context.Context, teamID string, in DomainContactInput) (*DomainContact, error) {
	var out DomainContact
	return &out, c.do(ctx, http.MethodPost, teamDomainContactsPath(teamID), in, &out)
}

func (c *Client) ListDomainContacts(ctx context.Context, teamID string) ([]DomainContact, error) {
	var out []DomainContact
	return out, c.do(ctx, http.MethodGet, teamDomainContactsPath(teamID), nil, &out)
}

func (c *Client) GetDomainContact(ctx context.Context, teamID, name string) (*DomainContact, error) {
	var out DomainContact
	return &out, c.do(ctx, http.MethodGet, teamDomainContactsPath(teamID)+"/"+url.PathEscape(name), nil, &out)
}

func (c *Client) UpdateDomainContact(ctx context.Context, teamID, name string, in DomainContactInput) (*DomainContact, error) {
	var out DomainContact
	return &out, c.do(ctx, http.MethodPatch, teamDomainContactsPath(teamID)+"/"+url.PathEscape(name), in, &out)
}

func (c *Client) DeleteDomainContact(ctx context.Context, teamID, name string) error {
	return c.do(ctx, http.MethodDelete, teamDomainContactsPath(teamID)+"/"+url.PathEscape(name), nil, nil)
}

func (c *Client) SetDefaultDomainContact(ctx context.Context, teamID, name string) error {
	return c.do(ctx, http.MethodPost, teamDomainContactsPath(teamID)+"/"+url.PathEscape(name)+"/set-default", nil, nil)
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

func (c *Client) ConnectDomain(ctx context.Context, teamID, projectID, name string) (*ConnectDomainResponse, error) {
	var out ConnectDomainResponse
	body := map[string]any{"name": name}
	return &out, c.do(ctx, http.MethodPost, teamDomainsPath(teamID, projectID)+"/connect", body, &out)
}

func (c *Client) VerifyDomain(ctx context.Context, teamID, projectID, name string) (*VerifyDomainResponse, error) {
	var out VerifyDomainResponse
	return &out, c.do(ctx, http.MethodPost, teamDomainsPath(teamID, projectID)+"/"+url.PathEscape(name)+"/verify", nil, &out)
}

func (c *Client) DisconnectDomain(ctx context.Context, teamID, projectID, name string) error {
	return c.do(ctx, http.MethodDelete, teamDomainsPath(teamID, projectID)+"/"+url.PathEscape(name), nil, nil)
}

