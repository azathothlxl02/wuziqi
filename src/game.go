package src

import (
	"image/color"
	"math"
	"os"

	"GUI/utils"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type Game struct {
	board       [BoardSize][BoardSize]Stone
	currentTurn Stone
	winner      Stone
	moves       int
	state       GameState
	playMode    PlayMode
	moveHistory [][2]int
	pendingAI   bool
}

func NewGame() *Game {
	utils.InitFont()
	return &Game{
		state: StateModeSelect,
	}
}

func (g *Game) Reset(mode PlayMode) {
	g.board = [BoardSize][BoardSize]Stone{}
	g.currentTurn = Black
	g.winner = Empty
	g.moves = 0
	g.playMode = mode
	g.state = StatePlaying
}

func (g *Game) Update() error {
	switch g.state {
	case StateModeSelect:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			_, y := ebiten.CursorPosition()

			centerY := WindowHeight / 2
			spacing := 60
			itemHeight := 32
			startY := centerY - spacing*2 + spacing*1

			switch {
			case y >= startY && y < startY+itemHeight:
				g.Reset(HumanVsHuman)
			case y >= startY+spacing && y < startY+spacing+itemHeight:
				g.Reset(HumanVsAI)
			case y >= startY+2*spacing && y < startY+2*spacing+itemHeight:
				os.Exit(0)
			}
		}

	case StatePlaying:
		if g.pendingAI {
			row, col := GetMCTMove(g.board)
			g.placeStoneAt(row, col)
			g.pendingAI = false
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StateModeSelect
			return nil
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			if x >= WindowWidth-120 && y >= WindowHeight-50 {
				g.undoMove()
				return nil
			}

			if g.playMode == HumanVsAI && g.currentTurn == White {
				return nil
			}

			g.handlePlayerMove()
		}
	case StateGameOver:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateModeSelect
		}
	}
	return nil
}

func (g *Game) handlePlayerMove() {
	x, y := ebiten.CursorPosition()
	col := int(math.Round((float64(x) - Margin) / TileSize))
	row := int(math.Round((float64(y) - Margin) / TileSize))
	if row >= 0 && row < BoardSize && col >= 0 && col < BoardSize && g.board[row][col] == Empty {
		g.placeStoneAt(row, col)
	}
}

func (g *Game) placeStoneAt(row, col int) {
	g.board[row][col] = g.currentTurn
	g.moveHistory = append(g.moveHistory, [2]int{row, col})
	g.moves++

	if g.checkWin(row, col) {
		g.winner = g.currentTurn
		g.state = StateGameOver
	} else if g.moves == BoardSize*BoardSize {
		g.winner = Empty
		g.state = StateGameOver
	} else {
		g.currentTurn = 3 - g.currentTurn

		if g.playMode == HumanVsAI && g.currentTurn == White {
			g.pendingAI = true
		}
	}
}

func (g *Game) checkWin(row, col int) bool {
	dirs := [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}
	for _, d := range dirs {
		count := 1
		for i := 1; i < 5; i++ {
			r, c := row+d[0]*i, col+d[1]*i
			if r < 0 || r >= BoardSize || c < 0 || c >= BoardSize || g.board[r][c] != g.currentTurn {
				break
			}
			count++
		}
		for i := 1; i < 5; i++ {
			r, c := row-d[0]*i, col-d[1]*i
			if r < 0 || r >= BoardSize || c < 0 || c >= BoardSize || g.board[r][c] != g.currentTurn {
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

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(WoodColor)

	switch g.state {
	case StateModeSelect:
		g.drawModeSelect(screen)
	case StatePlaying:
		g.drawBoard(screen)
		g.drawStatus(screen)
	case StateGameOver:
		g.drawBoard(screen)
		g.drawGameOver(screen)
	}
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	for i := 0; i < BoardSize; i++ {
		pos := float64(Margin + i*TileSize)
		ebitenutil.DrawLine(screen, pos, Margin, pos, WindowWidth-Margin, color.Black)
		ebitenutil.DrawLine(screen, Margin, pos, WindowWidth-Margin, pos, color.Black)
	}
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if g.board[r][c] != Empty {
				cx := float64(Margin + c*TileSize)
				cy := float64(Margin + r*TileSize)
				col := color.Black
				if g.board[r][c] == White {
					col = color.White
				}
				ebitenutil.DrawCircle(screen, cx, cy, StoneRadius, col)
			}
		}
	}
}
func (g *Game) undoMove() {
	if g.state != StatePlaying || len(g.moveHistory) == 0 {
		return
	}
	steps := 1
	if g.playMode == HumanVsAI && len(g.moveHistory) >= 2 {
		steps = 2
	}

	for i := 0; i < steps; i++ {
		last := g.moveHistory[len(g.moveHistory)-1]
		g.board[last[0]][last[1]] = Empty
		g.moveHistory = g.moveHistory[:len(g.moveHistory)-1]
		g.moves--
		g.currentTurn = 3 - g.currentTurn
	}
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	turnText := "Current Turn: "
	cx := float64(text.BoundString(utils.MplusFont, turnText).Dx() + 40)
	cy := float64(WindowWidth + StatusHeight/2)
	text.Draw(screen, "Press ESC to return menu", utils.MplusFont, 20, WindowHeight-65, color.Black)

	text.Draw(screen, turnText, utils.MplusFont, 20, int(cy+10), color.Black)

	col := color.Black
	if g.currentTurn == White {
		col = color.White
	}
	ebitenutil.DrawCircle(screen, cx, cy, StoneRadius, col)

	btnX, btnY := WindowWidth-120, WindowHeight-50
	ebitenutil.DrawRect(screen, float64(btnX), float64(btnY), 100, 30, color.RGBA{180, 180, 180, 255})
	text.Draw(screen, "Undo", utils.MplusFont, btnX+20, btnY+22, color.Black)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, float64(WindowWidth), float64(WindowWidth), OverlayColor)

	msg := "It's a Tie!"
	if g.winner == Black {
		msg = "Black Wins!"
	} else if g.winner == White {
		msg = "White Wins!"
	}
	utils.DrawCenteredText(screen, msg, "Click to return to menu", utils.MplusFont, WindowWidth)
}

func (g *Game) drawModeSelect(screen *ebiten.Image) {
	centerX := WindowWidth / 2
	centerY := WindowHeight / 2
	spacing := 60
	itemHeight := 32

	menuItems := []string{
		"Gomoku",
		"[1] Human vs Human",
		"[2] Human vs AI",
		"[3] Exit",
	}

	for i, item := range menuItems {
		bounds := text.BoundString(utils.MplusFont, item)
		x := centerX - bounds.Dx()/2
		y := centerY - spacing*2 + i*spacing + itemHeight/2
		text.Draw(screen, item, utils.MplusFont, x, y, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}
