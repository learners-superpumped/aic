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

func TestSaveAndLoadOIDCFields(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)
	if err := Save(&Profile{
		Name:     "default",
		Issuer:   "https://auth.example.com",
		ClientID: "cli-123",
	}); err != nil {
		t.Fatal(err)
	}
	got, err := Load("default")
	if err != nil {
		t.Fatal(err)
	}
	if got.Issuer != "https://auth.example.com" || got.ClientID != "cli-123" {
		t.Fatalf("oidc fields not round-tripped: %+v", got)
	}
}

func TestSaveAndLoadAudienceScope(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)
	if err := Save(&Profile{Name: "default", AudienceScope: "urn:zitadel:iam:org:project:id:123:aud"}); err != nil {
		t.Fatal(err)
	}
	got, err := Load("default")
	if err != nil {
		t.Fatal(err)
	}
	if got.AudienceScope != "urn:zitadel:iam:org:project:id:123:aud" {
		t.Fatalf("audience_scope not round-tripped: %+v", got)
	}
}

func TestLoadOrDefaultFillsEmptyProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)
	p := LoadOrDefault("default")
	if p.APIEndpoint != DefaultAPIEndpoint {
		t.Fatalf("APIEndpoint: %q", p.APIEndpoint)
	}
	if p.Issuer != DefaultIssuer {
		t.Fatalf("Issuer: %q", p.Issuer)
	}
	if p.ClientID != DefaultClientID {
		t.Fatalf("ClientID: %q", p.ClientID)
	}
	if p.AudienceScope != DefaultAudienceScope {
		t.Fatalf("AudienceScope: %q", p.AudienceScope)
	}
}

func TestLoadOrDefaultPreservesOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)
	if err := Save(&Profile{Name: "default", APIEndpoint: "https://staging.example.com", Issuer: "https://auth.staging.example.com"}); err != nil {
		t.Fatal(err)
	}
	p := LoadOrDefault("default")
	if p.APIEndpoint != "https://staging.example.com" {
		t.Fatalf("APIEndpoint override lost: %q", p.APIEndpoint)
	}
	if p.Issuer != "https://auth.staging.example.com" {
		t.Fatalf("Issuer override lost: %q", p.Issuer)
	}
	if p.ClientID != DefaultClientID {
		t.Fatalf("ClientID default not applied: %q", p.ClientID)
	}
}

func TestProfileTeamRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIC_CONFIG_DIR", dir)

	in := &Profile{Name: "default", Team: "team_abc"}
	if err := Save(in); err != nil {
		t.Fatal(err)
	}
	out, err := Load("default")
	if err != nil {
		t.Fatal(err)
	}
	if out.Team != "team_abc" {
		t.Fatalf("Team round-trip: want team_abc, got %q", out.Team)
	}
}
