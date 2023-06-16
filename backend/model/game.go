package model

import (
	"backend/lcr"
	"fmt"
)

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
