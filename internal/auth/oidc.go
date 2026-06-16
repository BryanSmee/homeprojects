package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Authenticator wraps an OIDC provider and OAuth2 client. It targets standard
// OIDC IdPs such as Pocket ID.
type Authenticator struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth2   oauth2.Config
}

// IDClaims are the user identity claims read from the ID token.
type IDClaims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

// NewAuthenticator performs OIDC discovery against issuer and returns a ready
// authenticator.
func NewAuthenticator(ctx context.Context, issuer, clientID, clientSecret, redirectURL string, scopes []string) (*Authenticator, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery for %q: %w", issuer, err)
	}
	return &Authenticator{
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
		oauth2: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		},
	}, nil
}

// AuthCodeURL builds the provider authorization URL for the given state.
func (a *Authenticator) AuthCodeURL(state string) string {
	return a.oauth2.AuthCodeURL(state)
}

// Exchange swaps an authorization code for verified identity claims.
func (a *Authenticator) Exchange(ctx context.Context, code string) (IDClaims, error) {
	var claims IDClaims

	tok, err := a.oauth2.Exchange(ctx, code)
	if err != nil {
		return claims, fmt.Errorf("token exchange: %w", err)
	}
	rawID, ok := tok.Extra("id_token").(string)
	if !ok {
		return claims, fmt.Errorf("no id_token in token response")
	}
	idToken, err := a.verifier.Verify(ctx, rawID)
	if err != nil {
		return claims, fmt.Errorf("verify id_token: %w", err)
	}
	if err := idToken.Claims(&claims); err != nil {
		return claims, fmt.Errorf("decode id_token claims: %w", err)
	}
	if claims.Subject == "" {
		return claims, fmt.Errorf("id_token missing subject")
	}
	return claims, nil
}
