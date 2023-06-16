package controllers

import (
	"context"
	"fmt"
	"strings"

	"backend/db" // <-- add this

	"github.com/gofiber/fiber/v2"
)

// Initialize the Firebase Auth client

// AuthRequired is a middleware function that validates the Authorization header and verifies the token using Firebase admin SDK
// @Summary Authentication required
// @Description Middleware function that validates the Authorization header and verifies the token using Firebase admin SDK
// @Tags Authentication
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Success 200 {string} string "OK"
// @Failure 401 {object} ErrorResponse
// @Router / [get]
func AuthRequired() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Get token from header
		bearerToken := c.Get("Authorization")
		splitToken := strings.Split(bearerToken, "Bearer ")
		token := splitToken[1]

		// Verify the token using your Firebase admin SDK

		tokenInfo, err := db.AuthClient.VerifyIDToken(context.Background(), token) // <-- change this
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString(fmt.Sprintf("Invalid ID token: %v\n", err))
		}

		// Set the user ID to context
		c.Locals("user", tokenInfo.UID)

		// Call the next handler
		return c.Next()
	}
}
