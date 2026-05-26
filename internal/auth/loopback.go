package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/oauth2"
)

// LoopbackLogin runs the Authorization Code + PKCE flow using a localhost
// redirect. openBrowser is called with the authorization URL (injected for
// testability; production passes OpenBrowser).
func LoopbackLogin(ctx context.Context, oc *OIDCConfig, openBrowser func(string) error) (*TokenSet, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer ln.Close()

	cfg := oc.OAuth2
	cfg.RedirectURL = fmt.Sprintf("http://%s/callback", ln.Addr().String())

	state := randString()
	verifier := oauth2.GenerateVerifier()

	type result struct {
		code string
		err  error
	}
	resCh := make(chan result, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			resCh <- result{err: fmt.Errorf("authorization failed: %s", e)}
			fmt.Fprintln(w, "Login failed. You may close this window.")
			return
		}
		if q.Get("state") != state {
			resCh <- result{err: fmt.Errorf("state mismatch")}
			fmt.Fprintln(w, "Login failed. You may close this window.")
			return
		}
		resCh <- result{code: q.Get("code")}
		fmt.Fprintln(w, "Login successful. You may close this window and return to the terminal.")
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	fmt.Println("Opening your browser to sign in. If it does not open, visit:")
	fmt.Println("  " + authURL)
	if err := openBrowser(authURL); err != nil {
		fmt.Println("Could not open a browser automatically.")
	}

	var res result
	select {
	case res = <-resCh:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if res.err != nil {
		return nil, res.err
	}

	tok, err := cfg.Exchange(ctx, res.code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, err
	}
	return tokenSetFrom(tok), nil
}

func randString() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
