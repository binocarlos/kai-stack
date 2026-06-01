package server

import (
	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/binocarlos/kai-stack/api/pkg/types"
	"github.com/gofiber/fiber/v3"
)

func (apiServer *StackAPIServer) RegisterUserRoutes() {
	// Login endpoint - no authentication required (generates JWT token)
	apiServer.router.Post("/user/login", apiServer.Login)

	// User status endpoint - requires authentication (returns session summary)
	apiServer.router.Get("/user/status", apiServer.RequireAuth, apiServer.GetUserStatus)

	// Logout endpoint - requires authentication (terminates session)
	apiServer.router.Post("/user/logout", apiServer.RequireAuth, apiServer.Logout)
}

// Login authenticates against the local fixed password and returns an HS256
// JWT. This is the dev/CI path; production logins go through Supabase (the
// frontend obtains a Supabase token directly). Disabled when the local
// provider is turned off.
func (apiServer *StackAPIServer) Login(c fiber.Ctx) error {
	if !apiServer.cfg.Auth.LocalEnabled {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "local password login is disabled",
		})
	}

	req, err := getRequestData[types.LoginRequest](c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	if req.Password != apiServer.cfg.WebServer.FixedPassword {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Incorrect password",
		})
	}

	token, err := apiServer.generateJWT(config.FIXED_USER_ID, []string{"admin"})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate authentication token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.LoginResponse{
		Token: token,
	})
}

// GetUserStatus is an authenticated endpoint that returns the current user's session summary
// This demonstrates how to extract the session data that was populated by the auth middleware
func (apiServer *StackAPIServer) GetUserStatus(c fiber.Ctx) error {
	userID, ok := GetUserIDFromContext(c)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}
	email, _ := GetUserEmailFromContext(c)
	roles, _ := GetUserRolesFromContext(c)
	return c.Status(fiber.StatusOK).JSON(&types.UserStatusResponse{
		UserID: userID,
		Email:  email,
		Roles:  types.Roles(roles),
	})
}

// Logout is an authenticated endpoint that terminates the current user's session
func (apiServer *StackAPIServer) Logout(c fiber.Ctx) error {
	userID, ok := GetUserIDFromContext(c)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
