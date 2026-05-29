package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
)

func TestInvitesAcceptCmd_CallsAccept(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/invites/TOK/accept" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Team{ID: "team_1", Name: "Acme", Role: "member"})
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r}

	cmd := newInvitesAcceptCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"TOK"}); err != nil {
		t.Fatalf("accept: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Acme", "team_1", "member"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output: %s", want, out)
		}
	}
}

func TestInvitesShowCmd_CallsPreview(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/invites/TOK" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(api.InvitePreview{
			TeamName:       "Acme",
			Role:           "owner",
			InvitedByEmail: "alice@example.com",
		})
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r}

	cmd := newInvitesShowCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"TOK"}); err != nil {
		t.Fatalf("show: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Acme", "owner", "alice@example.com"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output: %s", want, out)
		}
	}
}
