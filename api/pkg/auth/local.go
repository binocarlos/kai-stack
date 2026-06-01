package auth

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ProviderLocal is the provider name for the local fixed-password path.
const ProviderLocal = "local"

// localClaims mirrors the claims that server.generateJWT issues for the local
// fixed-password login (HS256, signed with the configured JWT secret).
type localClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// LocalAuthenticator verifies the HS256 tokens issued by the local login
// endpoint. It exists for dev/CI so the stack is usable without Supabase.
type LocalAuthenticator struct {
	secret []byte
}

// NewLocalAuthenticator builds a verifier for tokens signed with secret.
func NewLocalAuthenticator(secret string) *LocalAuthenticator {
	return &LocalAuthenticator{secret: []byte(secret)}
}

func (a *LocalAuthenticator) Name() string { return ProviderLocal }

func (a *LocalAuthenticator) Authenticate(_ context.Context, tokenString string) (*Identity, error) {
	token, err := jwt.ParseWithClaims(tokenString, &localClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid local token")
	}
	claims, ok := token.Claims.(*localClaims)
	if !ok || claims.UserID == "" {
		return nil, fmt.Errorf("local token missing user_id")
	}

	roles := claims.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}
	return &Identity{
		Subject:  claims.UserID,
		Roles:    roles,
		Provider: ProviderLocal,
	}, nil
}
