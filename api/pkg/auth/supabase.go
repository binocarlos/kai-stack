package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// ProviderSupabase is the provider name for Supabase-issued tokens.
const ProviderSupabase = "supabase"

// supabaseClaims are the Supabase access-token claims we read. The subject
// (RegisteredClaims.Subject, the "sub" claim) is the auth.users UUID.
type supabaseClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// SupabaseAuthenticator verifies Supabase access tokens using the project's
// JWKS endpoint (asymmetric signing keys), validating issuer, audience and
// expiry. No shared secret lives in this app and keys rotate transparently.
type SupabaseAuthenticator struct {
	keyfunc jwt.Keyfunc
	parser  *jwt.Parser
}

// NewSupabaseAuthenticator builds a verifier for a Supabase project. supabaseURL
// is the project URL (https://<ref>.supabase.co); audience is the expected aud
// claim ("authenticated"). It fetches the JWKS up front and refreshes it in the
// background.
func NewSupabaseAuthenticator(supabaseURL, audience string) (*SupabaseAuthenticator, error) {
	base := strings.TrimRight(supabaseURL, "/")
	if base == "" {
		return nil, fmt.Errorf("supabase URL is required")
	}
	jwksURL := base + "/auth/v1/.well-known/jwks.json"
	issuer := base + "/auth/v1"

	k, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to load Supabase JWKS from %s: %w", jwksURL, err)
	}

	parser := jwt.NewParser(
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
	)

	return &SupabaseAuthenticator{
		keyfunc: k.Keyfunc,
		parser:  parser,
	}, nil
}

func (a *SupabaseAuthenticator) Name() string { return ProviderSupabase }

func (a *SupabaseAuthenticator) Authenticate(_ context.Context, tokenString string) (*Identity, error) {
	claims := &supabaseClaims{}
	token, err := a.parser.ParseWithClaims(tokenString, claims, a.keyfunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid supabase token")
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("supabase token missing subject")
	}

	return &Identity{
		Subject:  claims.Subject,
		Email:    claims.Email,
		Roles:    []string{"user"},
		Provider: ProviderSupabase,
	}, nil
}
