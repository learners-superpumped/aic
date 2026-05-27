package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/app"
)

func TestBillingCmdHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newBillingCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"add-card", "cards", "topup", "balance", "history"} {
		if !names[want] {
			t.Errorf("missing billing subcommand %q", want)
		}
	}
	if names["status"] {
		t.Error("status subcommand should have been removed")
	}
}

func TestBillingBalanceHitsTeamScopedPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"balance_nano":50000000000,"balance_usd":50}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_123"}
	bal := findSub(newBillingCmd(), "balance")
	bal.SetContext(ctxWithApp(t, a))
	if err := bal.RunE(bal, nil); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/teams/team_123/billing/balance" {
		t.Fatalf("path = %q, want team-scoped", gotPath)
	}
	if !strings.Contains(buf.String(), "50") {
		t.Fatalf("expected balance in output: %s", buf.String())
	}
}

func TestBillingBalanceRequiresTeam(t *testing.T) {
	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New("http://unused", "tok"), Out: r} // no Team
	bal := findSub(newBillingCmd(), "balance")
	bal.SetContext(ctxWithApp(t, a))
	if err := bal.RunE(bal, nil); err == nil {
		t.Fatal("expected RequireTeam error when no team is selected")
	}
}
