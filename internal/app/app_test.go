package app

import (
	"bytes"
	"context"
	"testing"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
)

func TestContextRoundTrip(t *testing.T) {
	a := &App{
		Client:  api.New("http://x", "tok"),
		Project: "p1",
	}
	ctx := NewContext(context.Background(), a)
	got, err := FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Project != "p1" {
		t.Fatalf("project mismatch: %q", got.Project)
	}
}

func TestFromContextMissing(t *testing.T) {
	if _, err := FromContext(context.Background()); err == nil {
		t.Fatal("expected error when app is absent")
	}
}

func TestRequireProject(t *testing.T) {
	out, _ := newTestRenderer()
	a := &App{Project: "", Out: out}
	if err := a.RequireProject(); err == nil {
		t.Fatal("expected error when no project set")
	}
	a.Project = "p1"
	if err := a.RequireProject(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func newTestRenderer() (*Renderer, *bytes.Buffer) {
	var buf bytes.Buffer
	r, _ := NewRenderer("table", &buf)
	return r, &buf
}
