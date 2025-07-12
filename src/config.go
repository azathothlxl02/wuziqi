package src

import "image/color"

const (
	BoardSize    = 15
	TileSize     = 40
	Margin       = TileSize
	BoardWidth   = TileSize * (BoardSize - 1)
	WindowWidth  = BoardWidth + Margin*2
	StatusHeight = 60
	WindowHeight = WindowWidth + StatusHeight
	StoneRadius  = float64(TileSize) / 2 * 0.9
)

var (
	WoodColor    = color.RGBA{R: 210, G: 180, B: 140, A: 255}
	OverlayColor = color.RGBA{R: 0, G: 0, B: 0, A: 128}
)

type Stone int

const (
	Empty Stone = iota
	Black
	White
)

type GameState int

const (
	StateTitle GameState = iota
	StateModeSelect
	StatePlaying
	StateGameOver
)

type PlayMode int

const (
	HumanVsHuman PlayMode = iota
	HumanVsAI
)
