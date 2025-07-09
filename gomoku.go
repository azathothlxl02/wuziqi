package main

import (
	"image/color"
	"log"
	"math"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// --- Constants ---
const (
	boardSize    = 15
	tileSize     = 40
	margin       = tileSize
	boardWidth   = tileSize * (boardSize - 1)
	windowWidth  = boardWidth + margin*2
	statusHeight = 60
	windowHeight = windowWidth + statusHeight
	stoneRadius  = float64(tileSize) / 2 * 0.9
)

// --- Game States ---
type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateGameOver
)

type Stone int

const (
	Empty Stone = iota
	Black
	White
)

var (
	mplusFont    font.Face
	woodColor    = color.RGBA{R: 210, G: 180, B: 140, A: 255}
	overlayColor = color.RGBA{R: 0, G: 0, B: 0, A: 128} // Semi-transparent black
)

// --- Initialization ---
func init() {
	fontData, err := os.ReadFile("MPLUS1p-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}
	mplusFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// --- Game Struct ---
type Game struct {
	board       [boardSize][boardSize]Stone
	currentTurn Stone
	winner      Stone
	moves       int
	state       GameState
}

func NewGame() *Game {
	g := &Game{}
	g.Reset()
	g.state = StateTitle // Start at the title screen
	return g
}

func (g *Game) Reset() {
	g.board = [boardSize][boardSize]Stone{}
	g.currentTurn = Black
	g.winner = Empty
	g.moves = 0
	g.state = StatePlaying // When reset, go directly to playing
}

// --- Game Loop ---
func (g *Game) Update() error {
	switch g.state {
	case StateTitle:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StatePlaying
		}
	case StatePlaying:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.placeStone()
		}
	case StateGameOver:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.Reset()
			g.state = StateTitle // Go back to title screen after a game
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Always draw the board and stones
	g.drawBoard(screen)

	// Draw UI elements based on game state
	switch g.state {
	case StateTitle:
		drawCenteredText(screen, "Gomoku - 五子棋", "Click to Start", mplusFont)
	case StatePlaying:
		g.drawStatus(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}

// --- Drawing Helpers ---

func (g *Game) drawBoard(screen *ebiten.Image) {
	screen.Fill(woodColor)
	// Draw grid lines
	for i := 0; i < boardSize; i++ {
		pos := float64(margin + i*tileSize)
		ebitenutil.DrawLine(screen, pos, margin, pos, windowWidth-margin, color.Black)
		ebitenutil.DrawLine(screen, margin, pos, windowWidth-margin, pos, color.Black)
	}
	// Draw stones
	for r := 0; r < boardSize; r++ {
		for c := 0; c < boardSize; c++ {
			if g.board[r][c] != Empty {
				cx := float64(margin + c*tileSize)
				cy := float64(margin + r*tileSize)
				stoneColor := color.Black
				if g.board[r][c] == White {
					stoneColor = color.White
				}
				ebitenutil.DrawCircle(screen, cx, cy, stoneRadius, stoneColor)
			}
		}
	}
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	turnText := "Current Turn: "
	cx := float64(text.BoundString(mplusFont, turnText).Dx() + 40)
	cy := float64(windowWidth + statusHeight/2)

	text.Draw(screen, turnText, mplusFont, 20, int(cy+10), color.Black)

	stoneColor := color.Black
	if g.currentTurn == White {
		stoneColor = color.White
	}
	ebitenutil.DrawCircle(screen, cx, cy, stoneRadius, stoneColor)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Darken the board
	ebitenutil.DrawRect(screen, 0, 0, float64(windowWidth), float64(windowWidth), overlayColor)

	msg := "It's a Tie!"
	if g.winner == Black {
		msg = "Black Wins!"
	} else if g.winner == White {
		msg = "White Wins!"
	}
	drawCenteredText(screen, msg, "Click to Restart", mplusFont)
}

// Helper to draw centered text
func drawCenteredText(screen *ebiten.Image, line1, line2 string, face font.Face) {
	y1 := windowWidth/2 - 20
	y2 := windowWidth/2 + 30

	bounds1 := text.BoundString(face, line1)
	x1 := (windowWidth - bounds1.Dx()) / 2
	text.Draw(screen, line1, face, x1, y1, color.White)

	bounds2 := text.BoundString(face, line2)
	x2 := (windowWidth - bounds2.Dx()) / 2
	text.Draw(screen, line2, face, x2, y2, color.White)
}

// --- Game Logic ---
func (g *Game) placeStone() {
	x, y := ebiten.CursorPosition()
	col := int(math.Round((float64(x) - margin) / tileSize))
	row := int(math.Round((float64(y) - margin) / tileSize))

	if row >= 0 && row < boardSize && col >= 0 && col < boardSize && g.board[row][col] == Empty {
		g.board[row][col] = g.currentTurn
		g.moves++
		if g.checkWin(row, col) {
			g.winner = g.currentTurn
			g.state = StateGameOver
		} else if g.moves == boardSize*boardSize {
			g.winner = Empty
			g.state = StateGameOver
		} else {
			g.currentTurn = 3 - g.currentTurn
		}
	}
}

func (g *Game) checkWin(row, col int) bool {
	dirs := [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}} // H, V, Diag\, Diag/
	for _, d := range dirs {
		count := 1
		// Check in the positive direction
		for i := 1; i < 5; i++ {
			r, c := row+d[0]*i, col+d[1]*i
			if r < 0 || r >= boardSize || c < 0 || c >= boardSize || g.board[r][c] != g.currentTurn {
				break
			}
			count++
		}
		// Check in the negative direction
		for i := 1; i < 5; i++ {
			r, c := row-d[0]*i, col-d[1]*i
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

// --- Main Function ---
func main() {
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Gomoku - 五子棋")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
