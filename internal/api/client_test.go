package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoSendsAuthHeaderAndDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("missing/incorrect auth header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"proj_1","name":"alpha"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	var out Project
	if err := c.do(context.Background(), http.MethodGet, "/v1/projects/proj_1", nil, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
	if out.ID != "proj_1" || out.Name != "alpha" {
		t.Fatalf("decode mismatch: %+v", out)
	}
}

func TestDoStructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"code":"no_payment_method","message":"add a card first"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	err := c.do(context.Background(), http.MethodPost, "/v1/x", nil, nil)
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	if apiErr.Status != 403 || apiErr.Code != "no_payment_method" {
		t.Fatalf("unexpected error: %+v", apiErr)
	}
}

// The backend's 402 insufficient-credit body carries extra structured fields
// (balance/required/shortfall/topup_hint) alongside the standard {code,message}.
// The client must still surface the human-readable message, not fall back to the
// opaque "request failed with status 402".
func TestDoRichInsufficientCreditErrorSurfacesMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte(`{"code":"insufficient_credit","message":"insufficient credit: balance $5.00, need $12.00 (short $7.00) — run ` + "`aic billing topup`" + `","balance":5000000000,"required":12000000000,"shortfall":7000000000,"topup_hint":"run ` + "`aic billing topup --amount <usd>`" + `"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	err := c.do(context.Background(), http.MethodPost, "/v1/x", nil, nil)
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	if apiErr.Status != 402 || apiErr.Code != "insufficient_credit" {
		t.Fatalf("unexpected error: %+v", apiErr)
	}
	if !strings.Contains(apiErr.Error(), "insufficient credit") || !strings.Contains(apiErr.Error(), "$7.00") {
		t.Fatalf("message not surfaced (got opaque fallback?): %q", apiErr.Error())
	}
}

func TestListProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/team_1/projects" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`[{"id":"p1","name":"alpha"},{"id":"p2","name":"beta"}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	got, err := c.ListProjects(context.Background(), "team_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Name != "beta" {
		t.Fatalf("unexpected: %+v", got)
	}
}


func TestRefreshOn401ThenRetry(t *testing.T) {
	xCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/x" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		xCalls++
		if xCalls == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"code":"expired","message":"token expired"}`))
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer newtok" {
			t.Errorf("retry did not use refreshed token: %q", got)
		}
		w.Write([]byte(`{"id":"p1","name":"alpha"}`))
	}))
	defer srv.Close()

	refreshed := false
	persisted := ""
	c := New(srv.URL, "oldtok").WithRefresh(
		func(ctx context.Context) (*Tokens, error) {
			refreshed = true
			return &Tokens{AccessToken: "newtok", RefreshToken: "newref"}, nil
		},
		func(tok *Tokens) { persisted = tok.AccessToken },
	)
	var out Project
	if err := c.do(context.Background(), http.MethodGet, "/v1/x", nil, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
	if !refreshed || persisted != "newtok" || out.Name != "alpha" {
		t.Fatalf("refresh wiring wrong: refreshed=%v persisted=%q out=%+v", refreshed, persisted, out)
	}
	if xCalls != 2 {
		t.Fatalf("expected 2 calls, got %d", xCalls)
	}
}

func TestFriendly401WithoutRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":"unauthorized","message":"bad token"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok") // no refresh configured
	err := c.do(context.Background(), http.MethodGet, "/v1/x", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "aic login") {
		t.Fatalf("expected 401 error mentioning `aic login`, got %v", err)
	}
}

func TestSessionTokenUnmarshal(t *testing.T) {
	var s Session
	body := `{"session_id":"s1","status":"completed","access_token":"a","refresh_token":"r"}`
	if err := json.Unmarshal([]byte(body), &s); err != nil {
		t.Fatal(err)
	}
	if s.Tokens == nil || s.AccessToken != "a" || s.RefreshToken != "r" {
		t.Fatalf("embedded tokens not unmarshaled: %+v", s)
	}
}

func TestListTeamsHitsTeamsPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`[{"id":"team_1","name":"acme","role":"owner"}]`))
	}))
	defer srv.Close()

	teams, err := New(srv.URL, "tok").ListTeams(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/teams" {
		t.Fatalf("path: want /v1/teams, got %s", gotPath)
	}
	if len(teams) != 1 || teams[0].ID != "team_1" || teams[0].Role != "owner" {
		t.Fatalf("teams: %+v", teams)
	}
}

func TestCreateTeamPostsName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams" {
			t.Errorf("want POST /v1/teams, got %s %s", r.Method, r.URL.Path)
		}
		var in map[string]string
		json.NewDecoder(r.Body).Decode(&in)
		if in["name"] != "personal" {
			t.Errorf(`want body {"name":"personal"}, got %v`, in)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"team_1","name":"personal","role":"owner"}`))
	}))
	defer srv.Close()

	team, err := New(srv.URL, "tok").CreateTeam(context.Background(), "personal")
	if err != nil || team.ID != "team_1" {
		t.Fatalf("create team: %+v err=%v", team, err)
	}
}

func TestListProjectsScopedToTeam(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`[{"id":"p1","name":"alpha"}]`))
	}))
	defer srv.Close()

	if _, err := New(srv.URL, "tok").ListProjects(context.Background(), "team_1"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/teams/team_1/projects" {
		t.Fatalf("path: want /v1/teams/team_1/projects, got %s", gotPath)
	}
}

func TestCreateInviteSendsPostBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/team_1/invites" {
			t.Errorf("want POST /v1/teams/team_1/invites, got %s %s", r.Method, r.URL.Path)
		}
		var in map[string]string
		json.NewDecoder(r.Body).Decode(&in)
		if in["email"] != "bob@example.com" || in["role"] != "member" {
			t.Errorf("unexpected body: %v", in)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"inv_1","email":"bob@example.com","role":"member","status":"pending"}`))
	}))
	defer srv.Close()

	inv, err := New(srv.URL, "tok").CreateInvite(context.Background(), "team_1", "bob@example.com", "member")
	if err != nil || inv.ID != "inv_1" || inv.Status != "pending" {
		t.Fatalf("create invite: %+v err=%v", inv, err)
	}
}

func TestListInvitesHitsTeamPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/teams/team_1/invites" {
			t.Errorf("want GET /v1/teams/team_1/invites, got %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`[{"id":"inv_1","email":"bob@example.com","role":"member","status":"pending"}]`))
	}))
	defer srv.Close()

	invites, err := New(srv.URL, "tok").ListInvites(context.Background(), "team_1")
	if err != nil || len(invites) != 1 || invites[0].ID != "inv_1" {
		t.Fatalf("list invites: %+v err=%v", invites, err)
	}
}

func TestRevokeInviteSendsDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/teams/team_1/invites/inv_1" {
			t.Errorf("want DELETE /v1/teams/team_1/invites/inv_1, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := New(srv.URL, "tok").RevokeInvite(context.Background(), "team_1", "inv_1"); err != nil {
		t.Fatalf("revoke invite: %v", err)
	}
}

func TestResendInvitePostsToInvitePath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/team_1/invites/inv_1/resend" {
			t.Errorf("want POST /v1/teams/team_1/invites/inv_1/resend, got %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"id":"inv_1","email":"bob@example.com","role":"member","status":"pending"}`))
	}))
	defer srv.Close()

	inv, err := New(srv.URL, "tok").ResendInvite(context.Background(), "team_1", "inv_1")
	if err != nil || inv.ID != "inv_1" {
		t.Fatalf("resend invite: %+v err=%v", inv, err)
	}
}

func TestPreviewInviteGetsTokenPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/invites/tok_abc" {
			t.Errorf("want GET /v1/invites/tok_abc, got %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"team_name":"acme","role":"member","invited_by_email":"alice@example.com"}`))
	}))
	defer srv.Close()

	p, err := New(srv.URL, "tok").PreviewInvite(context.Background(), "tok_abc")
	if err != nil || p.TeamName != "acme" || p.InvitedByEmail != "alice@example.com" {
		t.Fatalf("preview invite: %+v err=%v", p, err)
	}
}

func TestAcceptInvitePostsToTokenAccept(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/invites/tok_abc/accept" {
			t.Errorf("want POST /v1/invites/tok_abc/accept, got %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"id":"team_1","name":"acme","role":"member"}`))
	}))
	defer srv.Close()

	team, err := New(srv.URL, "tok").AcceptInvite(context.Background(), "tok_abc")
	if err != nil || team.ID != "team_1" || team.Role != "member" {
		t.Fatalf("accept invite: %+v err=%v", team, err)
	}
}

func TestListMembersHitsTeamMembersPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/teams/team_1/members" {
			t.Errorf("want GET /v1/teams/team_1/members, got %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`[{"user_sub":"sub_1","role":"owner"},{"user_sub":"sub_2","role":"member"}]`))
	}))
	defer srv.Close()

	members, err := New(srv.URL, "tok").ListMembers(context.Background(), "team_1")
	if err != nil || len(members) != 2 || members[0].UserSub != "sub_1" {
		t.Fatalf("list members: %+v err=%v", members, err)
	}
}

func TestRemoveMemberDeletesUserSubPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/teams/team_1/members/sub_1" {
			t.Errorf("want DELETE /v1/teams/team_1/members/sub_1, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := New(srv.URL, "tok").RemoveMember(context.Background(), "team_1", "sub_1"); err != nil {
		t.Fatalf("remove member: %v", err)
	}
}

func TestSetMemberRolePatchesWithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/teams/team_1/members/sub_1" {
			t.Errorf("want PATCH /v1/teams/team_1/members/sub_1, got %s %s", r.Method, r.URL.Path)
		}
		var in map[string]string
		json.NewDecoder(r.Body).Decode(&in)
		if in["role"] != "admin" {
			t.Errorf("unexpected body: %v", in)
		}
		w.Write([]byte(`{"user_sub":"sub_1","role":"admin"}`))
	}))
	defer srv.Close()

	m, err := New(srv.URL, "tok").SetMemberRole(context.Background(), "team_1", "sub_1", "admin")
	if err != nil || m.UserSub != "sub_1" || m.Role != "admin" {
		t.Fatalf("set member role: %+v err=%v", m, err)
	}
}
