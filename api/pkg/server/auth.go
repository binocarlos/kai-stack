package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// ContextKey type for storing values in context
type ContextKey string

const (
	JWTUserIDContextKey    ContextKey = "jwtUserID"
	JWTUserRolesContextKey ContextKey = "jwtUserRoles"
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

// validateJWT validates and parses a JWT token string, returning the user information
func (apiServer *StackAPIServer) validateJWT(tokenString string) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Make sure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(apiServer.cfg.WebServer.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Check if token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims from JWT token")
	}

	return claims, nil
}

// RequireAuth is a middleware that validates authentication using JWT token or session ID
func (apiServer *StackAPIServer) RequireAuth(c fiber.Ctx) error {
	// Log the incoming request
	log.Debug().
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("ip", c.IP()).
		Msg("Auth middleware: Processing request")

	// Try JWT authentication first (Authorization header or token query parameter)
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

	if tokenString != "" {
		log.Debug().
			Str("path", c.Path()).
			Msg("Auth middleware: Attempting JWT authentication")

		// Validate JWT token
		jwtUser, err := apiServer.validateJWT(tokenString)
		if err != nil {
			log.Warn().
				Err(err).
				Str("path", c.Path()).
				Msg("Auth middleware: Invalid JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired JWT token",
			})
		}

		log.Debug().
			Str("userID", jwtUser.UserID).
			Strs("roles", jwtUser.Roles).
			Str("path", c.Path()).
			Msg("Auth middleware: JWT authentication successful")

		// Store JWT user in context
		c.Locals(string(JWTUserIDContextKey), jwtUser.UserID)
		c.Locals(string(JWTUserRolesContextKey), jwtUser.Roles)

		return c.Next()
	}

	return fmt.Errorf("authentication not found")
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
