package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
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

func TestEnsureDefaultTeamCreatesWhenEmpty(t *testing.T) {
	posted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte(`[]`))
		case http.MethodPost:
			posted = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"team_new","name":"personal","role":"owner"}`))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	}))
	defer srv.Close()

	id, created, err := ensureDefaultTeam(t.Context(), api.New(srv.URL, "tok"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team_new" {
		t.Errorf("want id team_new, got %q", id)
	}
	if !created {
		t.Errorf("want created==true")
	}
	if !posted {
		t.Errorf("expected POST /v1/teams to be called")
	}
}

func TestEnsureDefaultTeamSelectsFirstWhenNoDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("POST must not be called when teams exist")
		}
		w.Write([]byte(`[{"id":"team_1","name":"a","role":"owner"},{"id":"team_2","name":"b","role":"member"}]`))
	}))
	defer srv.Close()

	id, created, err := ensureDefaultTeam(t.Context(), api.New(srv.URL, "tok"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team_1" {
		t.Errorf("want id team_1, got %q", id)
	}
	if created {
		t.Errorf("want created==false")
	}
}

func TestEnsureDefaultTeamNoopWhenDefaultSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("POST must not be called when teams exist")
		}
		w.Write([]byte(`[{"id":"team_1","name":"a","role":"owner"}]`))
	}))
	defer srv.Close()

	id, created, err := ensureDefaultTeam(t.Context(), api.New(srv.URL, "tok"), "team_x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "" {
		t.Errorf("want empty id, got %q", id)
	}
	if created {
		t.Errorf("want created==false")
	}
}
