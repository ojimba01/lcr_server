package model

import (
	"math/rand"
)

type Dice struct {
	Sides int   `json:"Sides"`
	Rolls []int `json:"Rolls,omitempty"`
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
