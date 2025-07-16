package src

import (
	"image/color"
	"math"
	"net"
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
	difficulty  DifficultyLevel
	moveHistory [][2]int
	pendingAI   bool
	conn        net.Conn
	role        string
	lanState    string
	lanIPs      []string
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
				g.state = StateDifficultySelect
			case y >= startY+2*spacing && y < startY+2*spacing+itemHeight:
				g.state = StateLANConnect
			case y >= startY+3*spacing && y < startY+3*spacing+itemHeight:
				os.Exit(0)
			}
		}

	case StateDifficultySelect:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			_, y := ebiten.CursorPosition()

			centerY := WindowHeight / 2
			spacing := 60
			itemHeight := 32
			startY := centerY - spacing*2 + spacing*1

			// "Easy", "Medium", "Hard" buttons
			if y >= startY && y < startY+itemHeight {
				g.difficulty = Easy
				g.Reset(HumanVsAI)
			} else if y >= startY+spacing && y < startY+spacing+itemHeight {
				g.difficulty = Medium
				g.Reset(HumanVsAI)
			} else if y >= startY+2*spacing && y < startY+2*spacing+itemHeight {
				g.difficulty = Hard
				g.Reset(HumanVsAI)
			}
		}

	case StatePlaying:
		if g.pendingAI {
			var row, col int // Declare variables to be used by either AI

			if g.difficulty == Hard {
				// On Hard, call your Python-based AlphaZero model
				row, col = GetAIMove(g.board)
			} else {
				// On Easy/Medium, call the regular MCTS AI
				row, col = GetMCTMove(g.board, g.currentTurn, g.difficulty)
			}
			// Place the stone returned by the chosen AI
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
		if g.playMode == HumanVsLAN {
			if (g.role == "host" && g.currentTurn == Black) ||
				(g.role == "client" && g.currentTurn == White) {
				g.handlePlayerMove()
			} else {
				go func() {
					if g.conn == nil {
						return
					}
					row, col, err := recvMove(g.conn)
					if err != nil {
						return
					}
					g.placeStoneAt(row, col)
				}()
			}
			return nil
		}

	case StateLANConnect:
		if len(g.lanIPs) == 0 {
			g.lanIPs = GetLocalIPs()
			g.lanState = "waiting"
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			g.lanState = "waiting"
			go func() {
				conn, err := hostGame()
				if err != nil {
					g.lanState = "failed"
					return
				}
				g.conn = conn
				g.role = "host"
				g.Reset(HumanVsLAN)
			}()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
			g.lanState = "waiting"
			go func() {
				conn, err := joinGame("127.0.0.1")
				if err != nil {
					g.lanState = "failed"
					return
				}
				g.conn = conn
				g.role = "client"
				g.Reset(HumanVsLAN)
			}()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StateModeSelect
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

	if g.playMode == HumanVsLAN && g.conn != nil {
		sendMove(g.conn, row, col)
	}
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

func (g *Game) drawDifficultySelect(screen *ebiten.Image) {
	centerX := WindowWidth / 2
	centerY := WindowHeight / 2
	spacing := 60
	itemHeight := 32

	menuItems := []string{
		"Select Difficulty",
		"[1] Easy",
		"[2] Medium",
		"[3] Hard",
	}

	for i, item := range menuItems {
		bounds := text.BoundString(utils.MplusFont, item)
		x := centerX - bounds.Dx()/2
		y := centerY - spacing*2 + i*spacing + itemHeight/2

		// Make the title a different color to distinguish it
		col := color.White
		text.Draw(screen, item, utils.MplusFont, x, y, col)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(WoodColor)

	switch g.state {
	case StateModeSelect:
		g.drawModeSelect(screen)
	case StateDifficultySelect:
		g.drawDifficultySelect(screen)
	case StateLANConnect:
		g.drawLANConnect(screen)
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
		"[3] LAN Battle",
		"[4] Exit",
	}

	for i, item := range menuItems {
		bounds := text.BoundString(utils.MplusFont, item)
		x := centerX - bounds.Dx()/2
		y := centerY - spacing*2 + i*spacing + itemHeight/2
		text.Draw(screen, item, utils.MplusFont, x, y, color.White)
	}
}

func (g *Game) drawLANConnect(screen *ebiten.Image) {
	title := "LAN Battle"
	tw := text.BoundString(utils.MplusFont, title).Dx()
	text.Draw(screen, title, utils.MplusFont, (WindowWidth-tw)/2, 100, color.White)

	y := 160
	text.Draw(screen, "Your LAN IPs:", utils.MplusFont, 80, y, color.White)
	y += 30
	for _, ip := range g.lanIPs {
		text.Draw(screen, "  - "+ip, utils.MplusFont, 100, y, color.White)
		y += 25
	}

	y += 40
	text.Draw(screen, "Press [H] to HOST", utils.MplusFont, 80, y, color.White)
	text.Draw(screen, "Press [J] to JOIN", utils.MplusFont, 80, y+35, color.White)
	text.Draw(screen, "ESC: Back to menu", utils.MplusFont, 80, y+70, color.Gray{150})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}
