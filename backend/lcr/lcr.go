package lcr

import (
	"fmt"
	"math/rand"
)

type LCRGame struct {
	Players  []*LCRPlayer
	Dice     *LCRDice
	Pot      int
	Turn     int
	Player   *LCRPlayer
	Winner   *LCRPlayer
	GameOver bool
}

func NewLCRGame(players []*LCRPlayer, rollResults []int) *LCRGame {
	return &LCRGame{
		Players:  players,
		Dice:     NewLCRDice(rollResults),
		Pot:      0,
		Turn:     0,
		Player:   players[0],
		Winner:   nil,
		GameOver: false,
	}
}

const minPlayers = 3

func (g *LCRGame) Play() error {
	if len(g.Players) < minPlayers {
		return fmt.Errorf("not enough players to start the game, minimum required: %d", minPlayers)
	}

	for !g.GameOver {
		g.Player = g.Players[g.Turn]
		g.Player.TakeTurn(g)
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
	return nil
}

type LCRPlayer struct {
	Name  string
	Chips int
}

func NewLCRPlayer(name string) *LCRPlayer {
	return &LCRPlayer{
		Name:  name,
		Chips: 3,
	}
}

func (p *LCRPlayer) TakeTurn(g *LCRGame) {
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
}

func (p *LCRPlayer) GiveChip(player *LCRPlayer) {
	p.Chips--
	player.Chips++
}

func (p *LCRPlayer) PutInPot(g *LCRGame) {
	p.Chips--
	g.Pot++
}

type LCRDice struct {
	Sides int
	Rolls []int
}

func NewLCRDice(rolls []int) *LCRDice {
	return &LCRDice{
		Sides: 6,
		Rolls: rolls,
	}
}

func (d *LCRDice) Roll(numDice int) []int {
	rolls := make([]int, numDice)
	for i := 0; i < numDice; i++ {
		roll := rand.Intn(d.Sides) + 1
		rolls[i] = roll
		d.Rolls = append(d.Rolls, roll)
	}
	return rolls
}

func (d *LCRDice) GetRolls() []int {
	return d.Rolls
}

// 3XW4LgX0jMeo6mwTU9NrE0a2rYN2
