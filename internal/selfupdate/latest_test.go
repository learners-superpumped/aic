package selfupdate

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestVersion_ParsesTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/learners-superpumped/aic/releases/latest" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(`{"tag_name":"v1.5.0"}`))
	}))
	defer srv.Close()

	old := releasesLatestURL
	releasesLatestURL = srv.URL + "/repos/learners-superpumped/aic/releases/latest"
	defer func() { releasesLatestURL = old }()

	got, err := LatestVersion(t.Context())
	if err != nil {
		t.Fatalf("LatestVersion: %v", err)
	}
	if got != "v1.5.0" {
		t.Errorf("got %q, want v1.5.0", got)
	}
}

func TestLatestVersion_ErrorsOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	old := releasesLatestURL
	releasesLatestURL = srv.URL
	defer func() { releasesLatestURL = old }()

	if _, err := LatestVersion(t.Context()); err == nil {
		t.Fatal("expected error on non-200, got nil")
	}
}
