package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newDiscoveryServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"issuer":                        srv.URL,
			"authorization_endpoint":        srv.URL + "/authorize",
			"token_endpoint":                srv.URL + "/oauth/token",
			"device_authorization_endpoint": srv.URL + "/oauth/device",
			"jwks_uri":                      srv.URL + "/keys",
		})
	})
	t.Cleanup(srv.Close)
	return srv
}

func TestDiscoverBuildsConfig(t *testing.T) {
	srv := newDiscoveryServer(t)
	oc, err := Discover(context.Background(), srv.URL, "client-123")
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if oc.OAuth2.ClientID != "client-123" {
		t.Fatalf("client id: %q", oc.OAuth2.ClientID)
	}
	if oc.OAuth2.Endpoint.AuthURL != srv.URL+"/authorize" {
		t.Fatalf("auth url: %q", oc.OAuth2.Endpoint.AuthURL)
	}
	if oc.OAuth2.Endpoint.TokenURL != srv.URL+"/oauth/token" {
		t.Fatalf("token url: %q", oc.OAuth2.Endpoint.TokenURL)
	}
	if oc.DeviceAuthURL != srv.URL+"/oauth/device" {
		t.Fatalf("device url: %q", oc.DeviceAuthURL)
	}
}
