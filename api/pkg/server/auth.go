package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/binocarlos/kai-stack/api/pkg/auth"
	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// ContextKey type for storing values in context
type ContextKey string

const (
	JWTUserIDContextKey    ContextKey = "jwtUserID"
	JWTUserRolesContextKey ContextKey = "jwtUserRoles"
	JWTUserEmailContextKey ContextKey = "jwtUserEmail"
)

// JWTClaims represents the claims stored in the JWT token
type JWTClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// generateJWT creates a JWT token for the given user information
func (apiServer *StackAPIServer) generateJWT(userID string, roles []string) (string, error) {
	// Create the claims
	claims := JWTClaims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "badcode-api",
			Subject:   userID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(apiServer.cfg.WebServer.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// buildAuthenticators constructs the enabled token verifiers in priority order
// (Supabase first, then local). At least one provider must be enabled.
func buildAuthenticators(cfg *config.Config) ([]auth.Authenticator, error) {
	var authenticators []auth.Authenticator

	if cfg.Auth.SupabaseEnabled {
		if cfg.Auth.SupabaseURL == "" {
			return nil, fmt.Errorf("AUTH_SUPABASE_ENABLED is set but SUPABASE_URL is empty")
		}
		supa, err := auth.NewSupabaseAuthenticator(cfg.Auth.SupabaseURL, cfg.Auth.SupabaseAud)
		if err != nil {
			return nil, fmt.Errorf("failed to set up Supabase authentication: %w", err)
		}
		authenticators = append(authenticators, supa)
	}

	if cfg.Auth.LocalEnabled {
		if cfg.WebServer.JWTSecret == "" {
			return nil, fmt.Errorf("AUTH_LOCAL_ENABLED is set but SERVER_JWT_SECRET is empty")
		}
		authenticators = append(authenticators, auth.NewLocalAuthenticator(cfg.WebServer.JWTSecret))
	}

	if len(authenticators) == 0 {
		return nil, fmt.Errorf("no authentication providers enabled: set AUTH_SUPABASE_ENABLED and/or AUTH_LOCAL_ENABLED")
	}
	return authenticators, nil
}

// RequireAuth is a middleware that verifies the bearer token against the
// configured providers, resolves the canonical app profile, and stores its id
// and roles in the request context for downstream handlers.
func (apiServer *StackAPIServer) RequireAuth(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	var tokenString string
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	}
	if tokenString == "" {
		tokenString = c.Query("token")
	}
	if tokenString == "" {
		tokenString = c.Query("access_token")
	}
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	ctx := c.Context()

	// Try each provider in order; the first to accept the token wins.
	var identity *auth.Identity
	for _, authenticator := range apiServer.authenticators {
		id, err := authenticator.Authenticate(ctx, tokenString)
		if err != nil {
			log.Debug().Err(err).Str("provider", authenticator.Name()).Str("path", c.Path()).
				Msg("Auth middleware: provider rejected token")
			continue
		}
		identity = id
		break
	}
	if identity == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired token",
		})
	}

	// Map the verified identity onto our own canonical user.
	profile, err := apiServer.store.Profiles().GetOrCreate(ctx, identity)
	if err != nil {
		log.Error().Err(err).Str("provider", identity.Provider).Msg("Auth middleware: failed to resolve profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to resolve user",
		})
	}

	c.Locals(string(JWTUserIDContextKey), profile.ID)
	c.Locals(string(JWTUserRolesContextKey), []string(profile.Roles))
	c.Locals(string(JWTUserEmailContextKey), profile.Email)

	return c.Next()
}

// GetJWTUserFromContext retrieves the JWT user from the request context
func GetUserIDFromContext(c fiber.Ctx) (string, bool) {
	jwtUser, ok := c.Locals(string(JWTUserIDContextKey)).(string)
	return jwtUser, ok
}

func GetUserRolesFromContext(c fiber.Ctx) ([]string, bool) {
	jwtUserRoles, ok := c.Locals(string(JWTUserRolesContextKey)).([]string)
	return jwtUserRoles, ok
}

func GetUserEmailFromContext(c fiber.Ctx) (string, bool) {
	email, ok := c.Locals(string(JWTUserEmailContextKey)).(string)
	return email, ok
}
