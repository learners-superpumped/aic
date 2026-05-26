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
	for _, want := range []string{"add-card", "cards", "status"} {
		if !names[want] {
			t.Errorf("missing billing subcommand %q", want)
		}
	}
}

func TestBillingCardsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"card_id":"c1","brand":"visa","last4":"4242","exp_month":12,"exp_year":2030,"default":true}]`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r}
	cards := findSub(newBillingCmd(), "cards")
	cards.SetContext(ctxWithApp(t, a))
	if err := cards.RunE(cards, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "4242") {
		t.Fatalf("expected last4 in output: %s", buf.String())
	}
}
