package auth

import (
	"context"
)

// DeviceLogin runs the OAuth 2.0 Device Authorization Grant. prompt is called
// with the verification URL and user code for the user to enter in a browser
// (injected for testability).
func DeviceLogin(ctx context.Context, oc *OIDCConfig, prompt func(verificationURI, userCode string)) (*TokenSet, error) {
	da, err := oc.OAuth2.DeviceAuth(ctx)
	if err != nil {
		return nil, err
	}
	uri := da.VerificationURIComplete
	if uri == "" {
		uri = da.VerificationURI
	}
	prompt(uri, da.UserCode)

	tok, err := oc.OAuth2.DeviceAccessToken(ctx, da)
	if err != nil {
		return nil, err
	}
	return tokenSetFrom(tok), nil
}
