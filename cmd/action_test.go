package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/app"
)

func TestPrintActionTable(t *testing.T) {
	buf := new(bytes.Buffer)
	r, err := app.NewRenderer("table", buf)
	if err != nil {
		t.Fatal(err)
	}
	a := &app.App{Out: r}
	if err := printAction(a, actionResult{Name: "api.x.com", Type: "CNAME", Status: "deleted"}, "Deleted api.x.com CNAME"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(buf.String()); got != "Deleted api.x.com CNAME" {
		t.Errorf("table line = %q, want %q", got, "Deleted api.x.com CNAME")
	}
}

func TestPrintActionJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	r, err := app.NewRenderer("json", buf)
	if err != nil {
		t.Fatal(err)
	}
	a := &app.App{Out: r}
	if err := printAction(a, actionResult{Name: "api.x.com", Type: "CNAME", Status: "deleted"}, "ignored in json"); err != nil {
		t.Fatal(err)
	}
	var got actionResult
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("not valid JSON: %v (%q)", err, buf.String())
	}
	if got != (actionResult{Name: "api.x.com", Type: "CNAME", Status: "deleted"}) {
		t.Errorf("decoded = %+v", got)
	}
}

func TestPrintActionJSONOmitsEmptyType(t *testing.T) {
	buf := new(bytes.Buffer)
	r, _ := app.NewRenderer("json", buf)
	a := &app.App{Out: r}
	if err := printAction(a, actionResult{Name: "acme", Status: "default"}, "acme is now the default."); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "type") {
		t.Errorf("empty type should be omitted, got %q", buf.String())
	}
}
