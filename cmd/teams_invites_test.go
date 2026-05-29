package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
)

func makeInviteJSON(id, email, role string) string {
	return `{"id":"` + id + `","email":"` + email + `","role":"` + role + `","expires_at":"2026-06-01T00:00:00Z"}`
}

func TestTeamsInviteCmd_CallsCreateInvite(t *testing.T) {
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/team_1/invites" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(makeInviteJSON("inv_1", "bob@example.com", "member")))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1"}

	cmd := newTeamsInviteCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"bob@example.com"}); err != nil {
		t.Fatalf("invite: %v", err)
	}
	if gotBody["email"] != "bob@example.com" {
		t.Errorf("want email bob@example.com, got %q", gotBody["email"])
	}
	if gotBody["role"] != "member" {
		t.Errorf("want role member, got %q", gotBody["role"])
	}
	out := buf.String()
	if !strings.Contains(out, "bob@example.com") || !strings.Contains(out, "inv_1") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestTeamsInvitesListCmd_Renders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/teams/team_1/invites" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`[` + makeInviteJSON("inv_1", "alice@example.com", "owner") + `]`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1"}

	cmd := newTeamsInvitesListCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("invites list: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ID", "EMAIL", "ROLE", "EXPIRES", "inv_1", "alice@example.com", "owner"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output: %s", want, out)
		}
	}
}

func TestTeamsInvitesRevokeCmd_CallsDelete(t *testing.T) {
	deleted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/teams/team_1/invites/inv_1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1"}

	cmd := newTeamsInvitesRevokeCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"inv_1"}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
	if !strings.Contains(buf.String(), "revoked") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestTeamsInvitesResendCmd_CallsPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/team_1/invites/inv_1/resend" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(makeInviteJSON("inv_2", "carol@example.com", "member")))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1"}

	cmd := newTeamsInvitesResendCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"inv_1"}); err != nil {
		t.Fatalf("resend: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "carol@example.com") || !strings.Contains(out, "inv_2") {
		t.Errorf("unexpected output: %s", out)
	}
}

