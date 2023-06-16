package routes

import (
	"backend/controllers"
	"backend/db"
	"backend/model"
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GameRoute func for describe group of game routes.

func GameRoutes(app *fiber.App) {
	// Register new GET endpoint.
	app.Get("/games/:gameID", controllers.AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received GET request for game:", c.Params("gameID"))
		start := time.Now()
		err := controllers.GetGame(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("GET request for game:", c.Params("gameID"), "completed in", elapsed)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString(err.Error())
		}
		return nil
	})

	app.Get("/availableGames", controllers.AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received GET request for available games")
		start := time.Now()
		err := controllers.GetAvailableGames(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("GET request for available games completed in", elapsed)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		return nil
	})

	app.Get("/games/id/:lobbyCode", controllers.AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received GET request for gameID:", c.Params("lobbyCode"))
		start := time.Now()
		gameID, err := controllers.GetGameIDByLobbyCode(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("GET request for gameID:", c.Params("lobbyCode"), "completed in", elapsed)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"gameID": gameID,
		})
	})

	// Register new POST endpoint.
	app.Post("/games", controllers.AuthRequired(), func(c *fiber.Ctx) error {

		fmt.Println("Received POST request for creating a game")
		start := time.Now()
		err := controllers.CreateGame(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for creating a game completed in", elapsed)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		return nil
	})

	app.Post("/games/:lobbyCode/join", controllers.AuthRequired(), func(c *fiber.Ctx) error {
		lobbyCode := c.Params("lobbyCode")
		fmt.Println("Received POST request for joining a game:", lobbyCode)
		start := time.Now()
		err := controllers.JoinGame(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for joining a game:", lobbyCode, "completed in", elapsed)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		gameID, err := controllers.GetGameIDByLobbyCode(c, db.DbClient)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"gameID": gameID,
		})
	})

	app.Post("/games/:lobbyCode/players/:playerName/ready", controllers.AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for setting player ready:", c.Params("playerName"), "in lobby:", c.Params("lobbyCode"))
		start := time.Now()
		err := controllers.SetPlayerReady(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for setting player ready:", c.Params("playerName"), "in lobby:", c.Params("lobbyCode"), "completed in", elapsed)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error setting player ready: %v\n", err))
		}
		return nil
	})

	app.Post("/games/:gameID/turn", controllers.AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for taking a turn in game:", c.Params("gameID"))
		start := time.Now()
		err := controllers.TakeTurn(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for taking a turn in game:", c.Params("gameID"), "completed in", elapsed)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error taking turn: %v\n", err))
		}
		return nil
	})

	app.Post("/games/:lobbyCode/addBots", func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for adding bots to game:", c.Params("lobbyCode"))
		start := time.Now()
		err := controllers.AddBotsToGame(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for adding bots to game:", c.Params("lobbyCode"), "completed in", elapsed)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error adding bots to game: %v\n", err))
		}
		return nil
	})

	// create the set bots to ready endpoint
	app.Post("/games/:lobbyCode/setBotsReady", func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for setting bots to ready in game:", c.Params("lobbyCode"))
		start := time.Now()
		err := controllers.SetBotsReady(c, db.DbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for setting bots to ready in game:", c.Params("lobbyCode"), "completed in", elapsed)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error setting bots to ready in game: %v\n", err))
		}
		return nil
	})

	app.Post("/games/:lobbyCode/start", controllers.AuthRequired(), func(c *fiber.Ctx) error {
		lobbyCode := c.Params("lobbyCode")
		fmt.Println("Received POST request for starting game:", lobbyCode)
		start := time.Now()

		// Find game with the given lobby code
		var game *model.Game
		gamesRef := db.DbClient.NewRef("games")
		if err := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).Get(context.Background(), &game); err != nil {
			return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("Error finding game with lobby code %s: %v\n", lobbyCode, err))
		}

		if err := game.Start(); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Error starting game: %v\n", err))
		}

		// Update the game state in the Firebase RTDB
		gameRef := gamesRef.Child(game.LobbyCode)
		if err := gameRef.Set(context.Background(), game); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error saving game to Firebase RTDB: %v\n", err))
		}

		elapsed := time.Since(start)
		fmt.Println("POST request for starting game:", lobbyCode, "completed in", elapsed)

		return c.JSON(fiber.Map{
			"game": game,
		})
	})

}
