package cmd

import (
	"testing"

	"github.com/learners-superpumped/aic/internal/app"
)

func TestResolveProjectFlagWins(t *testing.T) {
	got := resolveProject("flagproj", "defaultproj")
	if got != "flagproj" {
		t.Fatalf("want flagproj, got %q", got)
	}
}

func TestResolveProjectFallsBackToDefault(t *testing.T) {
	got := resolveProject("", "defaultproj")
	if got != "defaultproj" {
		t.Fatalf("want defaultproj, got %q", got)
	}
}

func TestBuildAppValidatesOutput(t *testing.T) {
	_, err := buildApp(buildAppArgs{
		profileName: "default",
		output:      "xml",
		apiEndpoint: "http://x",
		token:       "t",
	})
	if err == nil {
		t.Fatal("expected error for invalid output format")
	}
}

func TestBuildAppOK(t *testing.T) {
	a, err := buildApp(buildAppArgs{
		profileName:    "default",
		output:         "json",
		apiEndpoint:    "http://x",
		token:          "t",
		defaultProject: "p1",
		projectFlag:    "",
	})
	if err != nil {
		t.Fatal(err)
	}
	var _ *app.App = a
	if a.Project != "p1" {
		t.Fatalf("project: %q", a.Project)
	}
}

func TestRootCommandVersion(t *testing.T) {
	defer SetVersion(Version)
	SetVersion("9.9.9-test")
	cmd := NewRootCmd()
	if cmd.Version != "9.9.9-test" {
		t.Fatalf("expected version 9.9.9-test, got %q", cmd.Version)
	}
}
