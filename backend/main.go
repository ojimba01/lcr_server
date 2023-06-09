// Bare Bone Api Game of the LCR Game (Functional but not pretty)
package main

import (
	"backend/lcr"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"log"
	"math/rand"
	"time"

	// "github.com/joho/godotenv" // Keep commented out for Railway deployment

	_ "backend/docs"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	// swagger handler
	"github.com/gofiber/fiber/v2/middleware/cors"
	"google.golang.org/api/option"
)

// @Description LCR Game API

// client represents a Firebase auth client
var client *auth.Client

// Player represents a game player
type Player struct {
	Name        string `json:"Name"`
	Chips       int    `json:"Chips"`
	LobbyStatus bool   `json:"LobbyStatus"`
	UserID      string `json:"UserID,omitempty"`
}

// Game represents a game instance
type Game struct {
	Players   []*Player    `json:"Players"`
	Creator   *Player      `json:"Creator,omitempty"`
	Dice      *Dice        `json:"Dice,omitempty"`
	Pot       int          `json:"Pot"`
	Turn      int          `json:"Turn"`
	Player    *Player      `json:"Player,omitempty"`
	Winner    *Player      `json:"Winner,omitempty"`
	GameOver  bool         `json:"GameOver"`
	LobbyCode string       `json:"LobbyCode"`
	LCRGame   *lcr.LCRGame `json:"LCRGame,omitempty"`
	GameID    string       `json:"gameID,omitempty"`
}

// LoginData represents user login information
type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Dice represents a dice
type Dice struct {
	Sides int   `json:"Sides"`
	Rolls []int `json:"Rolls,omitempty"`
}

// GetAvailableGamesResponse represents the response structure for the available games endpoint
type GetAvailableGamesResponse struct {
	Games map[string]*Game `json:"games"`
}

// ErrorResponse represents the response structure for an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// LCRGames is a map that holds LCR games
var LCRGames = make(map[string]*lcr.LCRGame)

// FirebaseCredentials represents the structure of Firebase credentials
type FirebaseCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// CreateGameResponse represents the response structure for the create game endpoint
type CreateGameResponse struct {
	GameID    string  `json:"gameID"`
	LobbyCode string  `json:"lobbyCode"`
	Creator   *Player `json:"creator"`
}

var (
	// credentials holds the Firebase credentials
	credentials FirebaseCredentials
)

// generateLobbyCode generates a random lobby code
func generateLobbyCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 5
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, codeLength)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// convertToLCRPlayers converts []*Player to []*lcr.LCRPlayer
func convertToLCRPlayers(players []*Player) []*lcr.LCRPlayer {
	lcrPlayers := make([]*lcr.LCRPlayer, len(players))
	for i, player := range players {
		lcrPlayers[i] = &lcr.LCRPlayer{
			Name:  player.Name,
			Chips: player.Chips,
		}
	}
	return lcrPlayers
}

// NewGame creates a new game instance
func NewGame(players []*Player) (*Game, *lcr.LCRGame) {
	// Initialize player chips to 3
	for _, player := range players {
		player.Chips = 3
	}

	dice := NewDice()

	// Initialize LCRGame
	lcrPlayers := convertToLCRPlayers(players)
	lcrGame := lcr.NewLCRGame(lcrPlayers, nil)

	game := &Game{
		Players:  players,
		Dice:     dice,
		Pot:      0,
		Turn:     0,
		Player:   players[0],
		Winner:   nil,
		GameOver: false,
	}
	return game, lcrGame
}

// NewPlayer creates a new player instance
func NewPlayer(name string) *Player {
	return &Player{
		Name:        name,
		Chips:       3,
		LobbyStatus: false, // players are not ready when they join
	}
}

// Start starts the game
func (g *Game) Start() error {
	if len(g.Players) < 3 {
		return fmt.Errorf("not enough players to start the game, minimum required: 3")
	}

	// Convert Players to LCRPlayers
	lcrPlayers := convertToLCRPlayers(g.Players)

	// Initialize the Dice
	g.Dice = &Dice{} // Initialize with default settings

	// Start the game
	g.LCRGame = lcr.NewLCRGame(lcrPlayers, nil)
	err := g.LCRGame.Play()
	if err != nil {
		return err
	}

	return nil
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
func getAvailableGames(c *fiber.Ctx, dbClient *db.Client) error {
	gamesRef := dbClient.NewRef("games")

	var games map[string]*Game
	if err := gamesRef.Get(context.Background(), &games); err != nil {
		log.Printf("Failed to retrieve games from Firebase RTDB: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve games from Firebase RTDB",
		})
	}

	availableGames := make(map[string]*Game)
	for gameID, game := range games {
		if !game.GameOver {
			availableGames[gameID] = game
		}
	}

	return c.JSON(GetAvailableGamesResponse{Games: availableGames})
}

// PlayTurn plays a turn in the game
func (g *Game) PlayTurn() {
	g.Player = g.Players[g.Turn]
	g.Dice.Rolls = g.Player.TakeTurnWithoutInput(g) // Update to store the dice roll results
	g.Turn++
	if g.Turn == len(g.Players) {
		g.Turn = 0
	}

	remainingPlayers := 0
	for _, player := range g.Players {
		if player.Chips > 0 {
			remainingPlayers++
		}
	}
	if remainingPlayers == 1 {
		g.GameOver = true
		for _, player := range g.Players {
			if player.Chips > 0 {
				g.Winner = player
				break
			}
		}
	}
}

// TakeTurnWithoutInput simulates taking a turn without player input
func (p *Player) TakeTurnWithoutInput(g *Game) []int {
	numDice := p.Chips
	if numDice > 3 {
		numDice = 3
	}
	rolls := g.Dice.Roll(numDice)

	for _, roll := range rolls {
		switch roll {
		case 4:
			p.GiveChip(g.Players[(g.Turn-1+len(g.Players))%len(g.Players)])
		case 5:
			p.PutInPot(g)
		case 6:
			p.GiveChip(g.Players[(g.Turn+1)%len(g.Players)])
		default:
		}
	}

	return rolls // Return the dice roll results
}

func (p *Player) GiveChip(player *Player) {
	p.Chips--
	player.Chips++
}

func (p *Player) PutInPot(g *Game) {
	p.Chips--
	g.Pot++
}

func NewDice() *Dice {
	return &Dice{
		Sides: 6,
		Rolls: []int{},
	}
}

func (d *Dice) Roll(numDice int) []int {
	rolls := make([]int, numDice)
	for i := 0; i < numDice; i++ {
		rolls[i] = rand.Intn(d.Sides) + 1
	}
	return rolls
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
func createGame(c *fiber.Ctx, dbClient *db.Client) error {
	var players []*Player
	if err := c.BodyParser(&players); err != nil {
		log.Printf("Error parsing player data: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid player data",
		})
	}

	if players == nil {
		log.Println("Player data is empty")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Player data is empty",
		})
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid user ID",
			})
		}
	}

	// Create game with the provided players
	game, lcrGame := NewGame(players)

	// Set the creator of the game
	game.Creator = players[0]

	// Generate lobby code
	game.LobbyCode = generateLobbyCode()

	// Each player in lobby is set to not ready at start
	for _, player := range game.Players {
		player.LobbyStatus = false
	}

	// Save the game to the Firebase RTDB
	gameRef, err := dbClient.NewRef("games").Push(context.Background(), nil)
	if err != nil {
		log.Printf("Failed to create game reference in Firebase RTDB: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create game reference in Firebase RTDB",
		})
	}

	gameID := gameRef.Key
	game.GameID = gameID

	if err := gameRef.Set(context.Background(), game); err != nil {
		log.Printf("Failed to save game to Firebase RTDB: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save game to Firebase RTDB",
		})
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
func joinGame(c *fiber.Ctx, dbClient *db.Client) error {
	lobbyCode := c.Params("lobbyCode")

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query games from Firebase RTDB: " + err.Error(),
		})
	}

	if len(results) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	gameKey := results[0].Key()
	gameSnapshot := gamesRef.Child(gameKey)
	var game Game
	if err := gameSnapshot.Get(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve game from Firebase RTDB: " + err.Error(),
		})
	}

	var playerData struct {
		Name string `json:"Name"`
	}
	if err := c.BodyParser(&playerData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid player data",
		})
	}
	if playerData.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Player name is empty",
		})
	}

	// Assign the user ID to the player
	userID := c.Locals("user").(string)
	player := NewPlayer(playerData.Name)
	player.UserID = userID

	// Add new player to the game
	game.Players = append(game.Players, player)

	// Save the game with the new player back to Firebase RTDB
	if err := gameSnapshot.Set(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save updated game to Firebase RTDB",
		})
	}

	return c.JSON(game)
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
func addBotsToGame(c *fiber.Ctx, dbClient *db.Client) error {
	lobbyCode := c.Params("lobbyCode")

	// Generate a random number of bots between 2 and 4
	numBots := rand.Intn(3) + 2

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query games from Firebase RTDB :( " + err.Error(),
		})
	}

	if len(results) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	gameKey := results[0].Key()
	gameSnapshot := gamesRef.Child(gameKey)
	var game Game
	if err := gameSnapshot.Get(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve game from Firebase RTDB" + err.Error(),
		})
	}

	// Add the new bots to the game
	for i := 0; i < numBots; i++ {
		botName := fmt.Sprintf("Bot %d", len(game.Players))
		bot := NewPlayer(botName)
		bot.UserID = "3XW4LgX0jMeo6mwTU9NrE0a2rYN2"
		game.Players = append(game.Players, bot)
	}

	// Save the game with the new bots back to Firebase RTDB
	if err := gameSnapshot.Set(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save updated game to Firebase RTDB",
		})
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
func setBotsReady(c *fiber.Ctx, dbClient *db.Client) error {
	lobbyCode := c.Params("lobbyCode")

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query games from Firebase RTDB :( " + err.Error(),
		})
	}

	if len(results) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	gameKey := results[0].Key()
	gameSnapshot := gamesRef.Child(gameKey)
	var game Game
	if err := gameSnapshot.Get(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve game from Firebase RTDB" + err.Error(),
		})
	}

	// Set everyone to ready
	for _, player := range game.Players {
		player.LobbyStatus = true
	}

	// Save the game with the new bots back to Firebase RTDB
	if err := gameSnapshot.Set(context.Background(), &game); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save updated game to Firebase RTDB",
		})
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
func takeTurn(c *fiber.Ctx, dbClient *db.Client) error {
	gameID := c.Params("gameID")
	gameRef := dbClient.NewRef("games/" + gameID)

	// Get the game from the Firebase RTDB
	game := &Game{}
	if err := gameRef.Get(context.Background(), game); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	if game.GameOver {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "The game has already ended",
		})
	}

	game.PlayTurn()

	// Update the game in the Firebase RTDB
	err := gameRef.Set(context.Background(), game)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save game to Firebase RTDB",
		})
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
func getGame(c *fiber.Ctx, dbClient *db.Client) error {
	gameID := c.Params("gameID")
	gameRef := dbClient.NewRef("games/" + gameID)

	// Get the game from the Firebase RTDB
	game := &Game{}
	if err := gameRef.Get(context.Background(), game); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	return c.JSON(fiber.Map{
		"game": game,
	})
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
func getGameIDByLobbyCode(c *fiber.Ctx, dbClient *db.Client) (string, error) {
	lobbyCode := c.Params("lobbyCode")

	// Find game with the given lobby code
	gamesRef := dbClient.NewRef("games")
	query := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).LimitToFirst(1)
	results, err := query.GetOrdered(context.Background())
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("Game not found")
	}

	gameKey := results[0].Key()

	return gameKey, nil
}

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
func setPlayerReady(c *fiber.Ctx, dbClient *db.Client) error {
	// Parse player name and lobby code from the URL parameters
	playerName := c.Params("playerName")

	// Find the game with the given lobby code
	gameID, err := getGameIDByLobbyCode(c, dbClient)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	gameRef := dbClient.NewRef("games/" + gameID)

	// Get the game from the Firebase RTDB
	game := &Game{}
	if err := gameRef.Get(context.Background(), game); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	// Find the player and set the lobby status to true
	for _, player := range game.Players {
		if player.Name == playerName {
			player.LobbyStatus = true
			break
		}
	}

	// Update the game in the Firebase RTDB
	if err := gameRef.Set(context.Background(), game); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save game to Firebase RTDB",
		})
	}

	return c.JSON(game)
}

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
		tokenInfo, err := client.VerifyIDToken(context.Background(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Set the user ID to context
		c.Locals("user", tokenInfo.UID)

		// Call the next handler
		return c.Next()
	}
}

func main() {
	// Load environment variables from .env file || Keep commented out for Railway deployment
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Failed to load environment variables: %v", err)
	// }

	// Connect to the PostgreSQL database
	// Use environment variables to get postgress password.
	var postgresPassword string
	if postgresPassword = os.Getenv("POSTGRES_PASSWORD"); postgresPassword == "" {
		log.Fatalf("Failed to get POSTGRES_PASSWORD environment variable")
	}

	db, err := sql.Open("postgres", fmt.Sprintf("postgresql://postgres:%s@containers-us-west-111.railway.app:6862/railway", postgresPassword))
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Retrieve the Firebase credentials from the database
	var jsonVal []byte
	err = db.QueryRow("SELECT firebase FROM credentials").Scan(&jsonVal)
	if err != nil {
		log.Fatalf("Failed to retrieve Firebase credentials: %v", err)
	}

	// Convert the JSON string into the FirebaseCredentials struct
	err = json.Unmarshal(jsonVal, &credentials)
	if err != nil {
		log.Fatalf("Failed to unmarshal Firebase credentials: %v", err)
	}

	// Convert the FirebaseCredentials struct back into a JSON object
	optBytes, err := json.Marshal(credentials)
	if err != nil {
		log.Fatalf("Failed to marshal Firebase credentials: %v", err)
	}

	// Create Firebase options with the retrieved credentials
	opt := option.WithCredentialsJSON(optBytes)

	// Initialize the Firebase App
	fbApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}
	// Initialize the Firebase Auth client
	client, err = fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth client: %v", err)
	}

	// Initialize the Firebase RTDB client
	dbClient, err := fbApp.DatabaseWithURL(context.Background(), "https://lcr-webapp-default-rtdb.firebaseio.com/")
	if err != nil {
		log.Fatalf("Failed to initialize Firebase RTDB client: %v", err)
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	app := fiber.New()
	app.Static("/", "static")
	app.Get("/swagger/*", swagger.HandlerDefault) // default

	// Allow all origins and methods
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}))

	app.Get("/games/:gameID", func(c *fiber.Ctx) error {
		fmt.Println("Received GET request for game:", c.Params("gameID"))
		start := time.Now()
		err := getGame(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("GET request for game:", c.Params("gameID"), "completed in", elapsed)
		return err
	})

	app.Get("/available-games", AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received GET request for available games")
		start := time.Now()
		err := getAvailableGames(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("GET request for available games completed in", elapsed)
		return err
	})

	app.Post("/games", AuthRequired(), func(c *fiber.Ctx) error {

		fmt.Println("Received POST request for creating a game")
		start := time.Now()
		err := createGame(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for creating a game completed in", elapsed)
		return err
	})

	app.Post("/games/:lobbyCode/join", AuthRequired(), func(c *fiber.Ctx) error {
		lobbyCode := c.Params("lobbyCode")
		fmt.Println("Received POST request for joining a game:", lobbyCode)
		start := time.Now()
		err := joinGame(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for joining a game:", lobbyCode, "completed in", elapsed)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to join the game. " + err.Error(),
			})
		}

		gameID, err := getGameIDByLobbyCode(c, dbClient)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve the game ID. " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"gameID": gameID,
		})
	})

	app.Get("/game-lobby/:lobbyCode", AuthRequired(), func(c *fiber.Ctx) error {
		lobbyCode := c.Params("lobbyCode")
		fmt.Println("Received GET request for game lobby:", lobbyCode)
		// Fetch the game lobby data using the lobby code
		// Return the game lobby data as the response
		// ...

		return c.JSON(fiber.Map{
			"lobbyCode": lobbyCode,
			// Include other game lobby data in the response as needed
		})
	})
	app.Get("/games/id/:lobbyCode", AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received GET request for gameID:", c.Params("lobbyCode"))
		start := time.Now()
		gameID, err := getGameIDByLobbyCode(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("GET request for gameID:", c.Params("lobbyCode"), "completed in", elapsed)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve the game ID. " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"gameID": gameID,
		})
	})

	app.Post("/games/:lobbyCode/players/:playerName/ready", AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for setting player ready:", c.Params("playerName"), "in lobby:", c.Params("lobbyCode"))
		start := time.Now()
		err := setPlayerReady(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for setting player ready:", c.Params("playerName"), "in lobby:", c.Params("lobbyCode"), "completed in", elapsed)
		return err
	})

	app.Post("/games/:gameID/turn", AuthRequired(), func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for taking a turn in game:", c.Params("gameID"))
		start := time.Now()
		err := takeTurn(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for taking a turn in game:", c.Params("gameID"), "completed in", elapsed)
		return err
	})

	app.Post("/games/:lobbyCode/addBots", func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for adding bots to game:", c.Params("lobbyCode"))
		start := time.Now()
		err := addBotsToGame(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for adding bots to game:", c.Params("lobbyCode"), "completed in", elapsed)
		return err
	})

	// create the set bots to ready endpoint
	app.Post("/games/:lobbyCode/setBotsReady", func(c *fiber.Ctx) error {
		fmt.Println("Received POST request for setting bots to ready in game:", c.Params("lobbyCode"))
		start := time.Now()
		err := setBotsReady(c, dbClient)
		elapsed := time.Since(start)
		fmt.Println("POST request for setting bots to ready in game:", c.Params("lobbyCode"), "completed in", elapsed)
		return err
	})

	app.Post("/games/:lobbyCode/start", AuthRequired(), func(c *fiber.Ctx) error {
		lobbyCode := c.Params("lobbyCode")
		fmt.Println("Received POST request for starting game:", lobbyCode)
		start := time.Now()

		// Find game with the given lobby code
		var game *Game
		gamesRef := dbClient.NewRef("games")
		if err := gamesRef.OrderByChild("LobbyCode").EqualTo(lobbyCode).Get(context.Background(), &game); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Game not found",
			})
		}

		if err := game.Start(); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Update the game state in the Firebase RTDB
		gameRef := gamesRef.Child(game.LobbyCode)
		if err := gameRef.Set(context.Background(), game); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save game to Firebase RTDB",
			})
		}

		elapsed := time.Since(start)
		fmt.Println("POST request for starting game:", lobbyCode, "completed in", elapsed)

		return c.JSON(fiber.Map{
			"game": game,
		})
	})

	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
