package routes

import (
	"github.com/gofiber/fiber/v2"
)

// SwaggerRoute func for descibe group of API Docs routes.
func StaticRoutes(app *fiber.App) {
	app.Static("/", "static")
}
