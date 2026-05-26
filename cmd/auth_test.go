package cmd

import (
	"testing"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/config"
)

func TestNewAuthCmdsRegistersAll(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newAuthCmds() {
		names[c.Name()] = true
	}
	for _, want := range []string{"login", "logout", "whoami", "configure"} {
		if !names[want] {
			t.Errorf("missing auth command %q", want)
		}
	}
}

func TestLoginRequiresIssuerConfigured(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)
	root := NewRootCmd()
	root.SetArgs([]string{"login", "--profile", "default"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error when issuer/client_id not configured")
	}
}

func TestLogoutRemovesProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)

	if err := config.Save(&config.Profile{
		Name:        "default",
		AccessToken: "tok",
		APIEndpoint: "http://x",
		Output:      "table",
	}); err != nil {
		t.Fatal(err)
	}
	root := NewRootCmd()
	root.SetArgs([]string{"logout", "--profile", "default"})
	if err := root.Execute(); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := config.Load("default"); err == nil {
		t.Fatal("expected profile to be gone after logout")
	}
}
