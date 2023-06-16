package responses

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse is a struct for error response
type ErrorResponse struct {
	Code     int         `json:"code,omitempty"`
	Message  string      `json:"message,omitempty"`
	Internal string      `json:"internal,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

// Define an error handling middleware
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Cast error to fiber.Error to retrieve details
	e, ok := err.(*fiber.Error)

	if ok {
		c.Status(e.Code)
		return c.JSON(&ErrorResponse{
			Code:     e.Code,
			Message:  e.Message,
			Internal: e.Error(),
			Data:     nil,
		})
	}

	// Fallback to status 500 and a generic error message if the error is not a fiber.Error
	c.Status(500)
	return c.JSON(&ErrorResponse{
		Code:     500,
		Message:  "An unexpected error occurred",
		Internal: err.Error(),
		Data:     nil,
	})
}
