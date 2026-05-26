package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestLoopbackLoginExchangesCode(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 srv.URL,
			"authorization_endpoint": srv.URL + "/authorize",
			"token_endpoint":         srv.URL + "/oauth/token",
			"jwks_uri":               srv.URL + "/keys",
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != "the-code" {
			t.Errorf("unexpected token request: %v", r.Form)
		}
		if r.Form.Get("code_verifier") == "" {
			t.Error("missing PKCE code_verifier")
		}
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

	openBrowser := func(authURL string) error {
		u, _ := url.Parse(authURL)
		if u.Query().Get("code_challenge_method") != "S256" {
			t.Errorf("expected S256 challenge method, got %q", u.Query().Get("code_challenge_method"))
		}
		if u.Query().Get("code_challenge") == "" {
			t.Error("missing code_challenge on auth URL")
		}
		redirect := u.Query().Get("redirect_uri")
		state := u.Query().Get("state")
		go func() {
			cb, _ := url.Parse(redirect)
			q := cb.Query()
			q.Set("code", "the-code")
			q.Set("state", state)
			cb.RawQuery = q.Encode()
			http.Get(cb.String())
		}()
		return nil
	}

	ts, err := LoopbackLogin(context.Background(), oc, openBrowser)
	if err != nil {
		t.Fatalf("LoopbackLogin: %v", err)
	}
	if ts.AccessToken != "acc" || ts.RefreshToken != "ref" || ts.IDToken != "idt" {
		t.Fatalf("unexpected tokens: %+v", ts)
	}
}
