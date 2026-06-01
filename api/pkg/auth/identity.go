// Package auth is the provider-agnostic authentication seam for the API. An
// Authenticator verifies a bearer token and returns a normalized Identity; the
// HTTP middleware then maps that Identity onto the app's own canonical user
// (a profiles row), so the rest of the app never depends on which provider
// authenticated the request. New providers are added by implementing
// Authenticator - nothing downstream changes.
package auth

import "context"

// Identity is the normalized result of verifying a token. Subject is the
// provider's stable user id (a Supabase auth.users UUID, or the local fixed
// user id). Roles are the default roles to assign when a profile is first
// created for this identity; once a profile exists the app owns its roles.
type Identity struct {
	Subject  string
	Email    string
	Roles    []string
	Provider string
}

// Authenticator verifies a single kind of bearer token.
type Authenticator interface {
	// Authenticate verifies the token and returns the identity it encodes.
	// It returns a non-nil error when the token is not valid for this provider.
	Authenticate(ctx context.Context, token string) (*Identity, error)
	// Name identifies the provider (e.g. "supabase", "local").
	Name() string
}
