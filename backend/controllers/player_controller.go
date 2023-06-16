package controllers

import (
	"context"
	// "backend/db"
	"firebase.google.com/go/v4/db"
	// "backend/errors"

	"backend/model"

	// "backend/util"

	"github.com/gofiber/fiber/v2"
)

// setPlayerReady sets the lobby status of a player to ready
// @Summary Set player ready status
// @Description Set the lobby status of a player to ready
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Param playerName path string true "Player name"
// @Success 200 {object} Game
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/:lobbyCode/players/:playerName/ready [post]
func SetPlayerReady(c *fiber.Ctx, dbClient *db.Client) error {
	playerName := c.Params("playerName")

	gameID, err := GetGameIDByLobbyCode(c, dbClient)
	if err != nil {
		// Return a fiber error with appropriate status code and message
		return fiber.NewError(fiber.StatusNotFound, "game not found")
	}

	gameRef := dbClient.NewRef("games/" + gameID)

	game := &model.Game{}
	if err := gameRef.Get(context.Background(), game); err != nil {
		// Return a fiber error with appropriate status code and message
		return fiber.NewError(fiber.StatusNotFound, "game not found")
	}

	for _, player := range game.Players {
		if player.Name == playerName {
			player.LobbyStatus = true
			break
		}
	}

	if err := gameRef.Set(context.Background(), game); err != nil {
		// Return a fiber error with appropriate status code and message
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save game to Firebase RTDB")
	}

	return c.JSON(game)
}
