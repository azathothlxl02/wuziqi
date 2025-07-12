package main

import (
	"log"

	"GUI/src"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(src.WindowWidth, src.WindowHeight)
	ebiten.SetWindowTitle("Gomoku - 五子棋")

	game := src.NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
