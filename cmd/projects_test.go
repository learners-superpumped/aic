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

// runProjectsList executes the `list` subcommand with an injected App over a test server.
func runProjectsList(t *testing.T, srvURL string) string {
	t.Helper()
	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srvURL, "tok"), Out: r}

	cmd := newProjectsCmd()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "list" {
			sub.SetContext(app.NewContext(t.Context(), a))
			if err := sub.RunE(sub, nil); err != nil {
				t.Fatalf("list: %v", err)
			}
		}
	}
	return buf.String()
}

func TestProjectsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":"p1","name":"alpha"}]`))
	}))
	defer srv.Close()

	out := runProjectsList(t, srv.URL)
	if !strings.Contains(out, "alpha") {
		t.Fatalf("expected alpha in output: %s", out)
	}
}

func TestProjectsCmdHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newProjectsCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"list", "create", "delete", "use", "show"} {
		if !names[want] {
			t.Errorf("missing projects subcommand %q", want)
		}
	}
}
