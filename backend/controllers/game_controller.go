package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	// "backend/db"
	"firebase.google.com/go/v4/db"
	// "backend/errors"
	"backend/lcr"

	"backend/model"
	// "backend/util"

	"github.com/gofiber/fiber/v2"
)

// GetAvailableGamesResponse represents the response structure for the available games endpoint
type GetAvailableGamesResponse struct {
	Games map[string]*model.Game `json:"games"`
}
type CreateGameResponse struct {
	GameID    string        `json:"gameID"`
	LobbyCode string        `json:"lobbyCode"`
	Creator   *model.Player `json:"creator"`
}

// LCRGames is a map that holds LCR games
var LCRGames = make(map[string]*lcr.LCRGame)

// generateLobbyCode generates a random lobby code
func GenerateLobbyCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 5
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, codeLength)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// addBotsToGame adds bots to the game
// @Summary Add bots to game
// @Description Adds a random number of bots (between 2 and 4) to the game identified by the provided lobby code in the Firebase Realtime Database
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Param lobbyCode path string true "Lobby code"
// @Success 200 {object} Game
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/add-bots/{lobbyCode} [put]
func AddBotsToGame(c *fiber.Ctx, dbClient *db.Client) error {
	lobbyCode := c.Params("lobbyCode")

	// Generate a random number of bots between 2 and 4
	numBots := rand.Intn(3) + 2

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query games from Firebase RTDB")
	}

	if len(results) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Game not found")
	}

	gameKey := results[0].Key()
	gameSnapshot := gamesRef.Child(gameKey)
	var game model.Game
	if err := gameSnapshot.Get(context.Background(), &game); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve game from Firebase RTDB")
	}

	// Add the new bots to the game
	for i := 0; i < numBots; i++ {
		botName := fmt.Sprintf("Bot %d", len(game.Players))
		bot := model.NewPlayer(botName)
		bot.UserID = "3XW4LgX0jMeo6mwTU9NrE0a2rYN2"
		game.Players = append(game.Players, bot)
	}

	// Save the game with the new bots back to Firebase RTDB
	if err := gameSnapshot.Set(context.Background(), &game); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save updated game to Firebase RTDB")
	}

	return c.JSON(game)
}

// setBotsReady sets all bots to ready in the game
// @Summary Set bots ready
// @Description Sets all the bots in the game identified by the provided lobby code in the Firebase Realtime Database to ready
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Param lobbyCode path string true "Lobby code"
// @Success 200 {object} Game
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/bots-ready/{lobbyCode} [put]
func SetBotsReady(c *fiber.Ctx, dbClient *db.Client) error {
	lobbyCode := c.Params("lobbyCode")

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to query games from Firebase RTDB")
	}

	if len(results) == 0 {
		return c.Status(fiber.StatusNotFound).SendString("Game not found")
	}

	gameKey := results[0].Key()
	gameSnapshot := gamesRef.Child(gameKey)
	var game model.Game
	if err := gameSnapshot.Get(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve game from Firebase RTDB")
	}

	// Set everyone to ready
	for _, player := range game.Players {
		player.LobbyStatus = true
	}

	// Save the game with the new bots back to Firebase RTDB
	if err := gameSnapshot.Set(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save updated game to Firebase RTDB")
	}

	return c.JSON(game)
}

// getGameIDByLobbyCode retrieves the game ID based on the lobby code
// @Summary Get game ID by lobby code
// @Description Retrieves the game ID based on the provided lobby code from the Firebase Realtime Database
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Param lobbyCode path string true "Lobby code"
// @Success 200 {string} string "OK"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/id/{lobbyCode} [get]
func GetGameIDByLobbyCode(c *fiber.Ctx, dbClient *db.Client) (string, error) {
	lobbyCode := c.Params("lobbyCode")

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "Failed to query games from Firebase RTDB")
	}

	if len(results) == 0 {
		return "", fiber.NewError(fiber.StatusNotFound, "Game not found")
	}

	gameKey := results[0].Key()

	return gameKey, nil
}

// getAvailableGames retrieves the list of available games
// @Summary Get available games
// @Description Retrieves the list of available games from the Firebase Realtime Database
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Success 200 {object} GetAvailableGamesResponse
// @Failure 500 {object} ErrorResponse
// @Router /available-games [get]
func GetAvailableGames(c *fiber.Ctx, dbClient *db.Client) error {
	gamesRef := dbClient.NewRef("games")

	var games map[string]*model.Game
	if err := gamesRef.Get(context.Background(), &games); err != nil {
		log.Printf("Failed to retrieve games from Firebase RTDB: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve games from Firebase RTDB",
		})
	}

	availableGames := make(map[string]*model.Game)
	for gameID, game := range games {
		if !game.GameOver {
			availableGames[gameID] = game
		}
	}

	return c.JSON(GetAvailableGamesResponse{Games: availableGames})
}

// CreateGame represents the request structure for the create game endpoint
// @Summary Create a new game
// @Description Create a new game with the provided players
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Success 200 {object} CreateGameResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games [post]
func CreateGame(c *fiber.Ctx, dbClient *db.Client) error {
	var players []*model.Player
	if err := c.BodyParser(&players); err != nil {
		log.Printf("Error parsing player data: %s", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid player data")
	}

	if players == nil {
		log.Println("Player data is empty")
		return c.Status(fiber.StatusBadRequest).SendString("Player data is empty")
	}

	// Attach the user ID to each player
	for _, player := range players {
		fmt.Println("Player:", player)
		if userID, ok := c.Locals("user").(string); ok {
			fmt.Println("User ID:", userID)
			player.UserID = userID
		} else {
			// Handle the case where the user ID is not a string
			log.Println("Invalid user ID")
			return c.Status(fiber.StatusBadRequest).SendString("Invalid user ID")
		}
	}

	// Create game with the provided players
	game, lcrGame := model.NewGame(players)

	// Set the creator of the game
	game.Creator = players[0]

	// Generate lobby code
	game.LobbyCode = GenerateLobbyCode()

	// Each player in lobby is set to not ready at start
	for _, player := range game.Players {
		player.LobbyStatus = false
	}

	// Save the game to the Firebase RTDB
	gameRef, err := dbClient.NewRef("games").Push(context.Background(), nil)
	if err != nil {
		log.Printf("Failed to create game reference in Firebase RTDB: %s", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create game reference in Firebase RTDB")
	}

	gameID := gameRef.Key
	game.GameID = gameID

	if err := gameRef.Set(context.Background(), game); err != nil {
		log.Printf("Failed to save game to Firebase RTDB: %s", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save game to Firebase RTDB")
	}

	// Initialize LCRGame
	LCRGames[gameID] = lcrGame

	return c.JSON(CreateGameResponse{
		GameID:    gameID,
		LobbyCode: game.LobbyCode,
		Creator:   game.Creator,
	})
}

// joinGame allows a player to join an existing game
// @Summary Join a game
// @Description Join an existing game with the provided lobby code
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Param lobbyCode path string true "Lobby code"
// @Success 200 {object} Game
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{lobbyCode}/join [post]
func JoinGame(c *fiber.Ctx, dbClient *db.Client) error {
	lobbyCode := c.Params("lobbyCode")

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to query games from Firebase RTDB")
	}

	if len(results) == 0 {
		return c.Status(fiber.StatusNotFound).SendString("Game not found")
	}

	gameKey := results[0].Key()
	gameSnapshot := gamesRef.Child(gameKey)
	var game model.Game
	if err := gameSnapshot.Get(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve game from Firebase RTDB")
	}

	var playerData struct {
		Name string `json:"Name"`
	}
	if err := c.BodyParser(&playerData); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid player data")
	}
	if playerData.Name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Player name is empty")
	}

	// Assign the user ID to the player
	userID := c.Locals("user").(string)
	player := model.NewPlayer(playerData.Name)
	player.UserID = userID

	// Add new player to the game
	game.Players = append(game.Players, player)

	// Save the game with the new player back to Firebase RTDB
	if err := gameSnapshot.Set(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save updated game to Firebase RTDB")
	}

	return c.JSON(game)
}

// takeTurn performs a player's turn in the game
// @Summary Perform player's turn
// @Description Takes a turn for the player in the game identified by the provided game ID in the Firebase Realtime Database
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Param gameID path string true "Game ID"
// @Success 200 {object} Game
// @Failure 404 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /games/{gameID}/turn [post]
func TakeTurn(c *fiber.Ctx, dbClient *db.Client) error {
	gameID := c.Params("gameID")
	gameRef := dbClient.NewRef("games/" + gameID)

	// Get the game from the Firebase RTDB
	game := &model.Game{}
	if err := gameRef.Get(context.Background(), game); err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Game not found")
	}

	if game.GameOver {
		return c.Status(fiber.StatusForbidden).SendString("Game is over")
	}

	game.PlayTurn()

	// Update the game in the Firebase RTDB
	err := gameRef.Set(context.Background(), game)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save updated game to Firebase RTDB")
	}

	return c.JSON(fiber.Map{
		"game": game,
	})
}

// getGame retrieves the game by game ID
// @Summary Get game by ID
// @Description Retrieves the game based on the provided game ID from the Firebase Realtime Database
// @Tags Games
// @Accept json
// @Produce json
// @Param c path string true "Fiber context"
// @Param dbClient path string true "Database client"
// @Param gameID path string true "Game ID"
// @Success 200 {object} Game
// @Failure 404 {object} ErrorResponse
// @Router /games/{gameID} [get]
func GetGame(c *fiber.Ctx, dbClient *db.Client) error {
	gameID := c.Params("gameID")
	gameRef := dbClient.NewRef("games/" + gameID)

	// Get the game from the Firebase RTDB
	game := &model.Game{}
	if err := gameRef.Get(context.Background(), game); err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Game not found")
	}

	return c.JSON(fiber.Map{
		"game": game,
	})
}
