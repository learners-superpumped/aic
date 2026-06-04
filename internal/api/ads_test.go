package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLaunchAdPostsToCorrectPath(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody AdLaunchRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"cmp_1","team_id":"team_1","project_id":"proj_1","provider":"meta","objective":"traffic","budget_type":"daily","budget_nano":5000000000,"status":"submitted","launch_token":"tok123","created_at":"2026-06-04T00:00:00Z","updated_at":"2026-06-04T00:00:00Z"}`))
	}))
	defer srv.Close()

	req := AdLaunchRequest{
		LaunchToken: "tok123",
		Objective:   "traffic",
		BudgetType:  "daily",
		BudgetNano:  5_000_000_000,
		StartAt:     time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC),
		Creative: AdCreative{
			StorageRef:     "bucket/key",
			Headline:       "Try AIC",
			Body:           "The platform for modern teams.",
			CTA:            "Learn More",
			DestinationURL: "https://runaic.com",
		},
	}

	c := New(srv.URL, "tok")
	got, err := c.LaunchAd(context.Background(), "team_1", "proj_1", req)
	if err != nil {
		t.Fatalf("LaunchAd: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method: want POST, got %s", gotMethod)
	}
	if gotPath != "/v1/teams/team_1/projects/proj_1/ads/campaigns" {
		t.Errorf("path: want /v1/teams/team_1/projects/proj_1/ads/campaigns, got %s", gotPath)
	}
	if got.ID != "cmp_1" || got.Status != "submitted" || got.LaunchToken != "tok123" {
		t.Errorf("response decode mismatch: %+v", got)
	}
	if gotBody.LaunchToken != "tok123" || gotBody.Objective != "traffic" {
		t.Errorf("request body mismatch: %+v", gotBody)
	}
}

func TestGetAdHitsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{"id":"cmp_2","team_id":"team_1","project_id":"proj_1","provider":"meta","objective":"awareness","budget_type":"lifetime","budget_nano":100000000000,"status":"active","launch_token":"tok456","created_at":"2026-06-04T00:00:00Z","updated_at":"2026-06-04T00:00:00Z"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	got, err := c.GetAd(context.Background(), "team_1", "proj_1", "cmp_2")
	if err != nil {
		t.Fatalf("GetAd: %v", err)
	}
	if gotPath != "/v1/teams/team_1/projects/proj_1/ads/campaigns/cmp_2" {
		t.Errorf("path: want .../ads/campaigns/cmp_2, got %s", gotPath)
	}
	if got.ID != "cmp_2" || got.Status != "active" {
		t.Errorf("decode mismatch: %+v", got)
	}
}

func TestListAdsHitsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`[{"id":"cmp_1","status":"active"},{"id":"cmp_2","status":"paused"}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	got, err := c.ListAds(context.Background(), "team_1", "proj_1")
	if err != nil {
		t.Fatalf("ListAds: %v", err)
	}
	if gotPath != "/v1/teams/team_1/projects/proj_1/ads/campaigns" {
		t.Errorf("path mismatch: %s", gotPath)
	}
	if len(got) != 2 || got[1].Status != "paused" {
		t.Errorf("list decode mismatch: %+v", got)
	}
}

func TestPauseAdPostsToCorrectPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := New(srv.URL, "tok").PauseAd(context.Background(), "team_1", "proj_1", "cmp_1"); err != nil {
		t.Fatalf("PauseAd: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/teams/team_1/projects/proj_1/ads/campaigns/cmp_1/pause" {
		t.Errorf("want POST .../pause, got %s %s", gotMethod, gotPath)
	}
}

func TestResumeAdPostsToCorrectPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := New(srv.URL, "tok").ResumeAd(context.Background(), "team_1", "proj_1", "cmp_1"); err != nil {
		t.Fatalf("ResumeAd: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/teams/team_1/projects/proj_1/ads/campaigns/cmp_1/resume" {
		t.Errorf("want POST .../resume, got %s %s", gotMethod, gotPath)
	}
}

func TestDeleteAdSendsDelete(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := New(srv.URL, "tok").DeleteAd(context.Background(), "team_1", "proj_1", "cmp_1"); err != nil {
		t.Fatalf("DeleteAd: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/v1/teams/team_1/projects/proj_1/ads/campaigns/cmp_1" {
		t.Errorf("want DELETE .../campaigns/cmp_1, got %s %s", gotMethod, gotPath)
	}
}
