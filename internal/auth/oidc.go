package auth

import (
	"context"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// TokenSet is the result of a successful OIDC login.
type TokenSet struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	Expiry       time.Time
}

// OIDCConfig holds the discovered endpoints and the oauth2 client config.
type OIDCConfig struct {
	OAuth2        oauth2.Config
	DeviceAuthURL string
	Provider      *oidc.Provider
}

// Discover performs OIDC discovery against issuer and returns a ready oauth2
// config for the given public clientID. RedirectURL is set per-flow by the
// loopback login.
func Discover(ctx context.Context, issuer, clientID string) (*OIDCConfig, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	var extra struct {
		DeviceAuthURL string `json:"device_authorization_endpoint"`
	}
	_ = provider.Claims(&extra)

	return &OIDCConfig{
		Provider:      provider,
		DeviceAuthURL: extra.DeviceAuthURL,
		OAuth2: oauth2.Config{
			ClientID: clientID,
			Endpoint: oauth2.Endpoint{
				AuthURL:       provider.Endpoint().AuthURL,
				TokenURL:      provider.Endpoint().TokenURL,
				DeviceAuthURL: extra.DeviceAuthURL,
			},
			Scopes: []string{oidc.ScopeOpenID, "profile", "email", oidc.ScopeOfflineAccess},
		},
	}, nil
}

func tokenSetFrom(tok *oauth2.Token) *TokenSet {
	ts := &TokenSet{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	}
	if id, ok := tok.Extra("id_token").(string); ok {
		ts.IDToken = id
	}
	return ts
}
