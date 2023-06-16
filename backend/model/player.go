package model

import (
	"backend/lcr"
)

// Player represents a game player
type Player struct {
	Name        string `json:"Name"`
	Chips       int    `json:"Chips"`
	LobbyStatus bool   `json:"LobbyStatus"`
	UserID      string `json:"UserID,omitempty"`
}

// NewPlayer creates a new player instance
func NewPlayer(name string) *Player {
	return &Player{
		Name:        name,
		Chips:       3,
		LobbyStatus: false, // players are not ready when they join
	}
}

// convertToLCRPlayers converts []*Player to []*lcr.LCRPlayer
func ConvertToLCRPlayers(players []*Player) []*lcr.LCRPlayer {
	lcrPlayers := make([]*lcr.LCRPlayer, len(players))
	for i, player := range players {
		lcrPlayers[i] = &lcr.LCRPlayer{
			Name:  player.Name,
			Chips: player.Chips,
		}
	}
	return lcrPlayers
}

func (p *Player) GiveChip(player *Player) {
	p.Chips--
	player.Chips++
}

func (p *Player) PutInPot(g *Game) {
	p.Chips--
	g.Pot++
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
