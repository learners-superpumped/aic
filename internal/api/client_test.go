package api

import (
	"context"
	"net/http"
	"net/http/httptest"
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
