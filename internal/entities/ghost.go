package entities

type Ghost struct {
	X, Y       float64
	CurrentDir Direction
	State      GhostState
}

type GhostState int

const (
	GhostNormal GhostState = iota
	GhostEaten
)
