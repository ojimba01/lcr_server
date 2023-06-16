package controllers

import (
	"backend/responses"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler represents a middleware for error handling.
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Create a custom response
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return c.Status(code).JSON(&responses.ErrorResponse{
			Code:     code,
			Message:  err.Error(),
			Internal: "",
		})
	}
}
