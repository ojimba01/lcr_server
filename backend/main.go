package main

import (
	"backend/controllers"
	"backend/db"
	"backend/routes"
	"backend/util"

	_ "github.com/lib/pq"

	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "backend/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// @title LCR API Documentation
// @version 1.0
// @description This is the API documentation for the LCR API. When you click on any endpoint, you can try out the API's functionality.
// @termsOfService http://swagger.io/terms/
// @contact.name Olayinka Jimba
// @contact.email ojimba01@gmail.com

// @host localhost:3000
// @BasePath /

func main() {

	util.LoadEnv()

	db.Init()

	// Create a channel to listen for OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s", sig)
		log.Println("Shutting down...")
		db.ClosePgConnection()
		os.Exit(0)
	}()

	rand.New(rand.NewSource(time.Now().UnixNano()))

	app := fiber.New(fiber.Config{
		ErrorHandler: controllers.ErrorHandler(),
	})

	app.Use(recover.New())
	app.Use(logger.New()) // <-- Add this line to use the Logger middleware.

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173,https://lcr.up.railway.app",
		AllowMethods: "GET,POST,PUT",
		AllowHeaders: "Origin, Content-Type, Accept, Bearer, Authorization",
	}))

	routes.GameRoutes(app)
	routes.SwaggerRoutes(app)
	routes.NotFoundRoute(app)
	routes.StaticRoutes(app)

	if err := app.Listen("0.0.0.0:3000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
