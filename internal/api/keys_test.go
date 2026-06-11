package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAPIKey_PostsBodyAndDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/team_1/keys" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["project_id"] != "proj_1" || body["expires_in_seconds"] != float64(3600) {
			t.Errorf("unexpected body: %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(APIKey{
			ID: "key_1", Prefix: "aic_sk_abcdef", Scopes: []string{"storage:read"},
			ProjectID: "proj_1", Status: "active", Key: "aic_sk_raw_secret",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	k, err := c.CreateAPIKey(context.Background(), "team_1", CreateAPIKeyRequest{
		Scopes: []string{"storage:read"}, ProjectID: "proj_1", ExpiresIn: 3600,
	})
	if err != nil {
		t.Fatal(err)
	}
	if k.ID != "key_1" || k.Key != "aic_sk_raw_secret" {
		t.Fatalf("decoded key = %+v", k)
	}
}

func TestListAPIKeys_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/team_1/keys" {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("project") != "proj_1" || q.Get("limit") != "10" || q.Get("cursor") != "abc" {
			t.Errorf("query = %v", q)
		}
		_ = json.NewEncoder(w).Encode(Page[APIKey]{Data: []APIKey{{ID: "key_1"}}, HasMore: false})
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	page, err := c.ListAPIKeys(context.Background(), "team_1", "proj_1", 10, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 || page.Data[0].ID != "key_1" {
		t.Fatalf("page = %+v", page)
	}
}

func TestRevokeAPIKey_Deletes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/teams/team_1/keys/key_1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := New(srv.URL, "tok").RevokeAPIKey(context.Background(), "team_1", "key_1"); err != nil {
		t.Fatal(err)
	}
}
