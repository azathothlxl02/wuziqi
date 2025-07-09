package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	boardSize  = 15
	tileSize   = 40
	margin     = tileSize / 2
	windowSize = boardSize * tileSize
)

type Stone int

const (
	Empty Stone = iota
	Black
	White
)

type Game struct {
	board       [boardSize][boardSize]Stone
	currentTurn Stone
	gameOver    bool
	winner      Stone
}

func NewGame() *Game {
	return &Game{
		currentTurn: Black,
	}
}

func (g *Game) Update() error {
	if g.gameOver {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.Reset()
		}
		return nil
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		col := int(math.Round(float64(x-margin) / float64(tileSize)))
		row := int(math.Round(float64(y-margin) / float64(tileSize)))
		if row >= 0 && row < boardSize && col >= 0 && col < boardSize {
			if g.board[row][col] == Empty {
				g.board[row][col] = g.currentTurn
				if g.checkWin(row, col) {
					g.gameOver = true
					g.winner = g.currentTurn
				} else {
					g.currentTurn = 3 - g.currentTurn // Switch turn
				}
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw background
	screen.Fill(color.RGBA{222, 184, 135, 255}) // light wood color

	// Draw grid lines
	for i := 0; i < boardSize; i++ {
		start := float64(margin)
		end := float64(windowSize - margin)
		pos := float64(margin + i*tileSize)
		ebitenutil.DrawLine(screen, start, pos, end, pos, color.Black)
		ebitenutil.DrawLine(screen, pos, start, pos, end, color.Black)
	}

	// Draw stones
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			cx := float64(margin + j*tileSize)
			cy := float64(margin + i*tileSize)
			if g.board[i][j] == Black {
				drawCircle(screen, cx, cy, 16, color.Black)
			} else if g.board[i][j] == White {
				drawCircle(screen, cx, cy, 16, color.White)
			}
		}
	}

	// Show game status
	if g.gameOver {
		msg := "Black wins!"
		if g.winner == White {
			msg = "White wins!"
		}
		ebitenutil.DebugPrintAt(screen, msg+" (Click to restart)", 10, windowSize-30)
	} else {
		turn := "Black"
		if g.currentTurn == White {
			turn = "White"
		}
		ebitenutil.DebugPrintAt(screen, "Current turn: "+turn, 10, windowSize-30)
	}
}

func drawCircle(screen *ebiten.Image, cx, cy float64, r float64, clr color.Color) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			x := cx + dx
			y := cy + dy
			if dx*dx+dy*dy <= r*r {
				screen.Set(int(x), int(y), clr)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowSize, windowSize + 40
}

func (g *Game) checkWin(row, col int) bool {
	dirs := [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}
	for _, d := range dirs {
		count := 1
		for i := 1; i < 5; i++ {
			r := row + d[0]*i
			c := col + d[1]*i
			if r < 0 || r >= boardSize || c < 0 || c >= boardSize || g.board[r][c] != g.currentTurn {
				break
			}
			count++
		}
		for i := 1; i < 5; i++ {
			r := row - d[0]*i
			c := col - d[1]*i
			if r < 0 || r >= boardSize || c < 0 || c >= boardSize || g.board[r][c] != g.currentTurn {
				break
			}
			count++
		}
		if count >= 5 {
			return true
		}
	}
	return false
}

func (g *Game) Reset() {
	g.board = [boardSize][boardSize]Stone{}
	g.currentTurn = Black
	g.gameOver = false
	g.winner = Empty
}

func main() {
	eg := NewGame()
	ebiten.SetWindowSize(windowSize, windowSize+40)
	ebiten.SetWindowTitle("Gomoku - 五子棋")
	if err := ebiten.RunGame(eg); err != nil {
		log.Fatal(err)
	}
}
