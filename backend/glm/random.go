package glm

import "math/rand/v2"

type Rand interface {
	IntN(n int) int
	Shuffle(n int, swap func(i, j int))
}

type globalRand struct{}

func (globalRand) IntN(n int) int                     { return rand.IntN(n) }
func (globalRand) Shuffle(n int, swap func(i, j int)) { rand.Shuffle(n, swap) }
