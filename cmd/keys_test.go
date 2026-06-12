package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
)

func TestParseExpiresIn(t *testing.T) {
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"", 0, false},
		{"90d", 90 * 86400, false},
		{"12h", 12 * 3600, false},
		{"30m", 1800, false},
		{"0d", 0, true},
		{"-5d", 0, true},
		{"banana", 0, true},
		{"90", 0, true},
	}
	for _, tt := range tests {
		got, err := parseExpiresIn(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseExpiresIn(%q): want error, got %d", tt.in, got)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("parseExpiresIn(%q) = (%d, %v), want %d", tt.in, got, err, tt.want)
		}
	}
}

func TestKeysCreateCmd_ScopedKey_PrintsRawOnce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/team_1/keys" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req api.CreateAPIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if len(req.Scopes) != 1 || req.Scopes[0] != "storage:read" || req.ProjectID != "proj_1" {
			t.Errorf("request = %+v", req)
		}
		if req.ExpiresIn != 90*86400 {
			t.Errorf("expires_in = %d", req.ExpiresIn)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(api.APIKey{
			ID: "key_1", Prefix: "aic_sk_abcdef", Scopes: req.Scopes,
			ProjectID: req.ProjectID, Status: "active", Key: "aic_sk_RAWSECRET",
		})
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "team_1", Project: "proj_1", Out: r}

	cmd := newKeysCreateCmd()
	cmd.Flags().String("project", "", "") // stand-in for the root persistent flag
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.Flags().Set("project", "proj_1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("scope", "storage:read"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("expires-in", "90d"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("create: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"key_1", "aic_sk_RAWSECRET", "will not be shown again"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
	if strings.Contains(out, "FULL-ACCESS") {
		t.Error("scoped key must not print the full-access banner")
	}
}

func TestKeysCreateCmd_ProjectNotSentUnlessFlagSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req api.CreateAPIKeyRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.ProjectID != "" {
			t.Errorf("project_id leaked from profile default: %q", req.ProjectID)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(api.APIKey{ID: "key_1", Key: "aic_sk_RAW", Status: "active"})
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	// Project set in the app (profile default) but --project NOT passed.
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "team_1", Project: "proj_default", Out: r}

	cmd := newKeysCreateCmd()
	cmd.Flags().String("project", "", "")
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.Flags().Set("scope", "teams:read"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("create: %v", err)
	}
}

func TestKeysCreateCmd_Star_Banner_AndProjectBinding(t *testing.T) {
	// team-bound "*": banner prints, no project sent
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req api.CreateAPIKeyRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Scopes) != 1 || req.Scopes[0] != "*" || req.ProjectID != "" {
			t.Errorf("team-bound request = %+v", req)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(api.APIKey{
			ID: "key_2", Scopes: []string{"*"}, Status: "active", Key: "aic_sk_STARRAW",
		})
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "team_1", Out: r}

	cmd := newKeysCreateCmd()
	cmd.Flags().String("project", "", "")
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.Flags().Set("scope", "*"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("team-bound star create: %v", err)
	}
	if !strings.Contains(buf.String(), "TEAM FULL-ACCESS") {
		t.Errorf("missing team danger banner:\n%s", buf.String())
	}

	// project-bound "*": project is sent, project banner shown
	var buf2 bytes.Buffer
	r2, _ := app.NewRenderer("table", &buf2)
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req api.CreateAPIKeyRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Scopes) != 1 || req.Scopes[0] != "*" || req.ProjectID != "proj_1" {
			t.Errorf("project-bound request = %+v", req)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(api.APIKey{
			ID: "key_3", Scopes: []string{"*"}, ProjectID: "proj_1", Status: "active", Key: "aic_sk_PROJSTAR",
		})
	}))
	defer srv2.Close()

	a2 := &app.App{Client: api.New(srv2.URL, "tok"), Team: "team_1", Project: "proj_1", Out: r2}
	cmd2 := newKeysCreateCmd()
	cmd2.Flags().String("project", "", "")
	cmd2.SetContext(app.NewContext(t.Context(), a2))
	cmd2.SetOut(&buf2)
	if err := cmd2.Flags().Set("scope", "*"); err != nil {
		t.Fatal(err)
	}
	if err := cmd2.Flags().Set("project", "proj_1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd2.RunE(cmd2, nil); err != nil {
		t.Fatalf("project-bound star create: %v", err)
	}
	if !strings.Contains(buf2.String(), "PROJECT FULL-ACCESS") {
		t.Errorf("missing project danger banner:\n%s", buf2.String())
	}
}

func TestKeysCreateCmd_Star_RejectsCombinedScopes(t *testing.T) {
	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New("http://unused", "tok"), Team: "team_1", Out: r}

	cmd := newKeysCreateCmd()
	cmd.Flags().String("project", "", "")
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.Flags().Set("scope", "*"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("scope", "storage:read"); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("want error for --scope '*' combined with another scope")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKeysListCmd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/teams/team_1/keys" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(api.Page[api.APIKey]{Data: []api.APIKey{
			{ID: "key_1", Prefix: "aic_sk_abcdef", Scopes: []string{"storage:read"}, ProjectID: "proj_1", Status: "active"},
		}})
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "team_1", Out: r}

	cmd := newKeysListCmd()
	cmd.Flags().String("project", "", "")
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("list: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"key_1", "aic_sk_abcdef", "storage:read", "active"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestKeysRevokeCmd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/teams/team_1/keys/key_1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "team_1", Out: r}

	cmd := newKeysRevokeCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"key_1"}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !strings.Contains(buf.String(), "revoked") {
		t.Errorf("missing confirmation:\n%s", buf.String())
	}
}
