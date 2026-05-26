package config

import (
	"os"
	"testing"
	"time"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)

	in := &Profile{
		Name:           "default",
		AccessToken:    "acc",
		RefreshToken:   "ref",
		ExpiresAt:      time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC),
		DefaultProject: "proj_abc",
		Output:         "table",
		APIEndpoint:    "https://api.aic.example.com",
	}
	if err := Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load("default")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.AccessToken != "acc" || got.RefreshToken != "ref" {
		t.Fatalf("tokens not round-tripped: %+v", got)
	}
	if got.DefaultProject != "proj_abc" || got.APIEndpoint != "https://api.aic.example.com" {
		t.Fatalf("config not round-tripped: %+v", got)
	}
	if !got.ExpiresAt.Equal(in.ExpiresAt) {
		t.Fatalf("expires_at mismatch: %v", got.ExpiresAt)
	}
}

func TestCredentialsFilePerms(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)
	if err := Save(&Profile{Name: "default", AccessToken: "x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(dir + "/credentials")
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("want 0600, got %o", info.Mode().Perm())
	}
}

func TestLoadMissingProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)
	if _, err := Load("nope"); err == nil {
		t.Fatal("expected error loading missing profile")
	}
}
