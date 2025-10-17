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

// Login authenticates a user with Carbon and returns a JWT token
func (apiServer *StackAPIServer) Login(c fiber.Ctx) error {
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
	return c.Status(fiber.StatusOK).JSON(&types.UserStatusResponse{
		UserID: userID,
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
