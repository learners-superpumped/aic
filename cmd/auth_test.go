package cmd

import (
	"testing"

	"github.com/learners-superpumped/aic/internal/config"
)

func TestAuthCmdRegistersSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newAuthCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"login", "logout", "status"} {
		if !names[want] {
			t.Errorf("missing auth subcommand %q", want)
		}
	}
	if names["whoami"] {
		t.Errorf("whoami should be replaced by status")
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
	root.SetArgs([]string{"auth", "logout", "--profile", "default"})
	if err := root.Execute(); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := config.Load("default"); err == nil {
		t.Fatal("expected profile to be gone after logout")
	}
}
