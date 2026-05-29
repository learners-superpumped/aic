package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
)

func makeMemberJSON(userSub, role, joinedAt string) string {
	return `{"user_sub":"` + userSub + `","role":"` + role + `","joined_at":"` + joinedAt + `"}`
}

func TestMembersListCmd_Renders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/teams/team_1/members" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`[` + makeMemberJSON("sub_abc", "owner", "2026-01-01") + `]`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1"}

	cmd := newMembersListCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("members list: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"USER_SUB", "ROLE", "JOINED", "sub_abc", "owner", "2026-01-01"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output: %s", want, out)
		}
	}
}

func TestMembersRemoveCmd_CallsDelete(t *testing.T) {
	deleted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/teams/team_1/members/sub_abc" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1"}

	cmd := newMembersRemoveCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"sub_abc"}); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
	if !strings.Contains(buf.String(), "removed") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestMembersSetRoleCmd_CallsPatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/teams/team_1/members/sub_abc" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(makeMemberJSON("sub_abc", "owner", "2026-01-01")))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Out: r, Team: "team_1"}

	cmd := newMembersSetRoleCmd()
	cmd.SetContext(app.NewContext(t.Context(), a))
	cmd.SetOut(&buf)
	if err := cmd.RunE(cmd, []string{"sub_abc", "owner"}); err != nil {
		t.Fatalf("set-role: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "sub_abc") || !strings.Contains(out, "owner") {
		t.Errorf("unexpected output: %s", out)
	}
}
