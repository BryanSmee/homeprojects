// Package auth handles SSO login via OIDC (e.g. Pocket ID), session cookies,
// and extracting the authenticated principal from a request.
package auth

import "context"

// Principal is the authenticated (or anonymous) caller derived from a request.
type Principal struct {
	UserID        string
	Subject       string
	Email         string
	Name          string
	Authenticated bool
}

// Anonymous is the principal used for unauthenticated requests.
var Anonymous = Principal{Authenticated: false}

type ctxKey struct{}

// WithPrincipal returns a copy of ctx carrying p.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// FromContext returns the principal stored in ctx, or Anonymous if none.
func FromContext(ctx context.Context) Principal {
	if p, ok := ctx.Value(ctxKey{}).(Principal); ok {
		return p
	}
	return Anonymous
}
