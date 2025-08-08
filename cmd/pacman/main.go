package main

import (
	"log"

	"pacman/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g := game.New()
	ebiten.SetWindowTitle("Pacman (Go + Ebiten)")
	ebiten.SetWindowResizable(false)
	ebiten.SetWindowSize(g.ScreenWidth(), g.ScreenHeight())
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
