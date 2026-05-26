package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoSendsAuthHeaderAndDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("missing/incorrect auth header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"proj_1","name":"alpha"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	var out Project
	if err := c.do(context.Background(), http.MethodGet, "/v1/projects/proj_1", nil, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
	if out.ID != "proj_1" || out.Name != "alpha" {
		t.Fatalf("decode mismatch: %+v", out)
	}
}

func TestDoStructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"code":"no_payment_method","message":"add a card first"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	err := c.do(context.Background(), http.MethodPost, "/v1/x", nil, nil)
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	if apiErr.Status != 403 || apiErr.Code != "no_payment_method" {
		t.Fatalf("unexpected error: %+v", apiErr)
	}
}

func TestListProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`[{"id":"p1","name":"alpha"},{"id":"p2","name":"beta"}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	got, err := c.ListProjects(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Name != "beta" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestCreateInboxPostsAddress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/p1/inboxes" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var in map[string]string
		json.NewDecoder(r.Body).Decode(&in)
		if in["address"] != "a@x.com" {
			t.Errorf("address not sent: %v", in)
		}
		w.Write([]byte(`{"address":"a@x.com","status":"active"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	got, err := c.CreateInbox(context.Background(), "p1", "a@x.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.Address != "a@x.com" || got.Status != "active" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestRefreshOn401ThenRetry(t *testing.T) {
	xCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/token/refresh":
			w.Write([]byte(`{"access_token":"newtok","refresh_token":"newref"}`))
		case "/v1/x":
			xCalls++
			if xCalls == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"code":"expired","message":"token expired"}`))
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer newtok" {
				t.Errorf("retry did not use refreshed token: %q", got)
			}
			w.Write([]byte(`{"id":"p1","name":"alpha"}`))
		}
	}))
	defer srv.Close()

	persisted := ""
	c := New(srv.URL, "oldtok").WithRefresh("oldref", func(tok *Tokens) { persisted = tok.AccessToken })
	var out Project
	if err := c.do(context.Background(), http.MethodGet, "/v1/x", nil, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
	if out.Name != "alpha" {
		t.Fatalf("expected retry to succeed: %+v", out)
	}
	if persisted != "newtok" {
		t.Fatalf("onRefresh not called with new token, got %q", persisted)
	}
	if xCalls != 2 {
		t.Fatalf("expected /v1/x to be called exactly twice (401 then retry), got %d", xCalls)
	}
}

func TestFriendly401WithoutRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":"unauthorized","message":"bad token"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok") // no refresh configured
	err := c.do(context.Background(), http.MethodGet, "/v1/x", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "aic login") {
		t.Fatalf("expected 401 error mentioning `aic login`, got %v", err)
	}
}

func TestSessionTokenUnmarshal(t *testing.T) {
	var s Session
	body := `{"session_id":"s1","status":"completed","access_token":"a","refresh_token":"r"}`
	if err := json.Unmarshal([]byte(body), &s); err != nil {
		t.Fatal(err)
	}
	if s.Tokens == nil || s.AccessToken != "a" || s.RefreshToken != "r" {
		t.Fatalf("embedded tokens not unmarshaled: %+v", s)
	}
}
