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

func TestDomainsBuyHitsTeamProjectPath(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"name":"foo.com","status":"registering","auto_renew":true}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1", Project: "proj_1"}
	buy := findSub(newDomainsCmd(), "buy")
	buy.SetContext(ctxWithApp(t, a))
	_ = buy.Flags().Set("years", "2")
	if err := buy.RunE(buy, []string{"foo.com"}); err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" || gotPath != "/v1/teams/team_1/projects/proj_1/domains" {
		t.Fatalf("path = %s %s", gotMethod, gotPath)
	}
	if !strings.Contains(buf.String(), "registering") {
		t.Fatalf("output: %s", buf.String())
	}
}

func TestDomainsBuyRequiresTeamAndProject(t *testing.T) {
	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New("http://unused", "tok"), Out: r, Team: "team_1"} // no Project
	buy := findSub(newDomainsCmd(), "buy")
	buy.SetContext(ctxWithApp(t, a))
	if err := buy.RunE(buy, []string{"foo.com"}); err == nil {
		t.Fatal("expected RequireProject error")
	}
}
