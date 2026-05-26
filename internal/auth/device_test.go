package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeviceLogin(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"issuer":                        srv.URL,
			"authorization_endpoint":        srv.URL + "/authorize",
			"token_endpoint":                srv.URL + "/oauth/token",
			"device_authorization_endpoint": srv.URL + "/oauth/device",
			"jwks_uri":                      srv.URL + "/keys",
		})
	})
	mux.HandleFunc("/oauth/device", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"device_code": "dev-1", "user_code": "WXYZ-1234",
			"verification_uri":          srv.URL + "/device",
			"verification_uri_complete": srv.URL + "/device?user_code=WXYZ-1234",
			"expires_in":                600, "interval": 1,
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "acc", "refresh_token": "ref", "id_token": "idt",
			"token_type": "Bearer", "expires_in": 3600,
		})
	})

	oc, err := Discover(context.Background(), srv.URL, "client-1")
	if err != nil {
		t.Fatal(err)
	}
	var shown, shownURI string
	prompt := func(verificationURI, userCode string) {
		shown = userCode
		shownURI = verificationURI
	}

	ts, err := DeviceLogin(context.Background(), oc, prompt)
	if err != nil {
		t.Fatalf("DeviceLogin: %v", err)
	}
	if ts.AccessToken != "acc" || ts.RefreshToken != "ref" {
		t.Fatalf("unexpected tokens: %+v", ts)
	}
	if shown != "WXYZ-1234" {
		t.Fatalf("user code not shown: %q", shown)
	}
	if shownURI != srv.URL+"/device?user_code=WXYZ-1234" {
		t.Fatalf("expected complete verification URI, got %q", shownURI)
	}
}

func TestDeviceLoginFallsBackToVerificationURI(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"issuer":                        srv.URL,
			"authorization_endpoint":        srv.URL + "/authorize",
			"token_endpoint":                srv.URL + "/oauth/token",
			"device_authorization_endpoint": srv.URL + "/oauth/device",
			"jwks_uri":                      srv.URL + "/keys",
		})
	})
	mux.HandleFunc("/oauth/device", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"device_code": "dev-1", "user_code": "WXYZ-1234",
			"verification_uri": srv.URL + "/device", "expires_in": 600, "interval": 1,
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "acc", "refresh_token": "ref", "id_token": "idt",
			"token_type": "Bearer", "expires_in": 3600,
		})
	})

	oc, err := Discover(context.Background(), srv.URL, "client-1")
	if err != nil {
		t.Fatal(err)
	}
	var shown, shownURI string
	prompt := func(verificationURI, userCode string) {
		shown = userCode
		shownURI = verificationURI
	}

	ts, err := DeviceLogin(context.Background(), oc, prompt)
	if err != nil {
		t.Fatalf("DeviceLogin: %v", err)
	}
	if ts.AccessToken != "acc" || ts.RefreshToken != "ref" {
		t.Fatalf("unexpected tokens: %+v", ts)
	}
	if shown != "WXYZ-1234" {
		t.Fatalf("user code not shown: %q", shown)
	}
	if shownURI != srv.URL+"/device" {
		t.Fatalf("expected fallback verification URI, got %q", shownURI)
	}
}
