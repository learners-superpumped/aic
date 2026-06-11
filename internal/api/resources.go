package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
)

// newIdempotencyKey returns a fresh random key for a single command invocation.
// One key per Topup call means the transport's transparent retries reuse it, so a
// timed-out/lost-response retry charges the card exactly once.
func newIdempotencyKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

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

func (c *Client) History(ctx context.Context, teamID string, limit int, cursor string) (Page[LedgerEntry], error) {
	path := listPath(teamBillingPath(teamID)+"/history", limit, cursor, nil)
	var out Page[LedgerEntry]
	return out, c.do(ctx, http.MethodGet, path, nil, &out)
}

func (c *Client) Usage(ctx context.Context, teamID, from, to string) (*UsageSummary, error) {
	q := url.Values{}
	if from != "" {
		q.Set("from", from)
	}
	if to != "" {
		q.Set("to", to)
	}
	path := teamBillingPath(teamID) + "/usage"
	if e := q.Encode(); e != "" {
		path += "?" + e
	}
	var out UsageSummary
	return &out, c.do(ctx, http.MethodGet, path, nil, &out)
}

func (c *Client) Topup(ctx context.Context, teamID string, amountCents int64) (*TopupResult, error) {
	var res TopupResult
	headers := http.Header{"Idempotency-Key": []string{newIdempotencyKey()}}
	return &res, c.doWithHeaders(ctx, http.MethodPost, teamBillingPath(teamID)+"/topup",
		map[string]int64{"amount_cents": amountCents}, &res, headers)
}

func (c *Client) GetAutoRecharge(ctx context.Context, teamID string) (*AutoRechargeConfig, error) {
	var cfg AutoRechargeConfig
	return &cfg, c.do(ctx, http.MethodGet, teamBillingPath(teamID)+"/auto-recharge", nil, &cfg)
}

func (c *Client) SetAutoRecharge(ctx context.Context, teamID string, in AutoRechargeInput) error {
	return c.do(ctx, http.MethodPut, teamBillingPath(teamID)+"/auto-recharge", in, nil)
}

// --- Teams ---

func (c *Client) ListTeams(ctx context.Context, limit int, cursor string) (Page[Team], error) {
	var out Page[Team]
	return out, c.do(ctx, http.MethodGet, listPath("/v1/teams", limit, cursor, nil), nil, &out)
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

func (c *Client) ListProjects(ctx context.Context, teamID string, limit int, cursor string) (Page[Project], error) {
	var out Page[Project]
	return out, c.do(ctx, http.MethodGet, listPath(teamProjectsPath(teamID), limit, cursor, nil), nil, &out)
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

func (c *Client) ListDomainContacts(ctx context.Context, teamID string, limit int, cursor string) (Page[DomainContact], error) {
	var out Page[DomainContact]
	return out, c.do(ctx, http.MethodGet, listPath(teamDomainContactsPath(teamID), limit, cursor, nil), nil, &out)
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

func (c *Client) ListDomains(ctx context.Context, teamID, projectID string, limit int, cursor string) (Page[Domain], error) {
	var out Page[Domain]
	return out, c.do(ctx, http.MethodGet, listPath(teamDomainsPath(teamID, projectID), limit, cursor, nil), nil, &out)
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

// --- Team invites ---

func teamInvitesPath(teamID string) string {
	return "/v1/teams/" + url.PathEscape(teamID) + "/invites"
}

func (c *Client) CreateInvite(ctx context.Context, teamID, email, role string) (*Invite, error) {
	var inv Invite
	body := map[string]string{"email": email, "role": role}
	return &inv, c.do(ctx, http.MethodPost, teamInvitesPath(teamID), body, &inv)
}

func (c *Client) ListInvites(ctx context.Context, teamID string, limit int, cursor string) (Page[Invite], error) {
	var out Page[Invite]
	return out, c.do(ctx, http.MethodGet, listPath(teamInvitesPath(teamID), limit, cursor, nil), nil, &out)
}

func (c *Client) RevokeInvite(ctx context.Context, teamID, id string) error {
	return c.do(ctx, http.MethodDelete, teamInvitesPath(teamID)+"/"+url.PathEscape(id), nil, nil)
}

func (c *Client) ResendInvite(ctx context.Context, teamID, id string) (*Invite, error) {
	var inv Invite
	return &inv, c.do(ctx, http.MethodPost, teamInvitesPath(teamID)+"/"+url.PathEscape(id)+"/resend", nil, &inv)
}

// --- Invite acceptance (token-based, team unknown until preview/accept) ---

func (c *Client) PreviewInvite(ctx context.Context, token string) (*InvitePreview, error) {
	var p InvitePreview
	return &p, c.do(ctx, http.MethodGet, "/v1/invites/"+url.PathEscape(token), nil, &p)
}

func (c *Client) AcceptInvite(ctx context.Context, token string) (*Team, error) {
	var t Team
	return &t, c.do(ctx, http.MethodPost, "/v1/invites/"+url.PathEscape(token)+"/accept", nil, &t)
}

// --- Team members ---

func teamMembersPath(teamID string) string {
	return "/v1/teams/" + url.PathEscape(teamID) + "/members"
}

func (c *Client) ListMembers(ctx context.Context, teamID string, limit int, cursor string) (Page[Member], error) {
	var out Page[Member]
	return out, c.do(ctx, http.MethodGet, listPath(teamMembersPath(teamID), limit, cursor, nil), nil, &out)
}

func (c *Client) RemoveMember(ctx context.Context, teamID, userSub string) error {
	return c.do(ctx, http.MethodDelete, teamMembersPath(teamID)+"/"+url.PathEscape(userSub), nil, nil)
}

func (c *Client) SetMemberRole(ctx context.Context, teamID, userSub, role string) (*Member, error) {
	var m Member
	return &m, c.do(ctx, http.MethodPatch, teamMembersPath(teamID)+"/"+url.PathEscape(userSub),
		map[string]string{"role": role}, &m)
}
