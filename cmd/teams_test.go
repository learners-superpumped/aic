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

func TestTeamsCmdHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newTeamsCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"list", "create", "switch", "show"} {
		if !names[want] {
			t.Errorf("missing teams subcommand %q", want)
		}
	}
}

func TestTeamsListRendersTeams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams" {
			t.Errorf("want /v1/teams, got %s", r.URL.Path)
		}
		w.Write([]byte(`[{"id":"team_1","name":"acme","role":"owner"}]`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r}

	cmd := newTeamsCmd()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "list" {
			sub.SetContext(app.NewContext(t.Context(), a))
			if err := sub.RunE(sub, nil); err != nil {
				t.Fatalf("list: %v", err)
			}
		}
	}
	if !strings.Contains(buf.String(), "acme") {
		t.Fatalf("expected acme in output: %s", buf.String())
	}
}
