package src

import (
	"fmt"
	"image/color"
	"math"
	"net"
	"os"
	"time"
	"wuziqi/utils"

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
	foundRooms  []RoomInfo
	selectedIdx int
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
			var row, col int

			if g.difficulty == Hard {
				row, col = GetAIMove(g.board)
			} else {
				row, col = GetMCTMove(g.board, g.currentTurn, g.difficulty)
			}
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
					row, col, err := recvMove(g.conn)
					if err == nil {
						g.placeStoneAt(row, col)
					}
				}()
			}
		}

	case StateLANConnect:
		if inpututil.IsKeyJustPressed(ebiten.KeyH) && g.conn == nil {
			g.lanState = "hosting"
			go func() {
				conn, err := HostGame()
				if err != nil {
					g.lanState = "failed"
					return
				}
				g.conn = conn
				g.role = "host"
				g.Reset(HumanVsLAN)
			}()
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			g.lanState = "searching"
			go func() {
				rooms, err := DiscoverRooms(2 * time.Second)
				if err != nil {
					g.lanState = "failed"
					return
				}
				g.foundRooms = rooms
				g.selectedIdx = 0
				g.lanState = "ready"
			}()
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.lanState == "ready" {
			if g.selectedIdx >= 0 && g.selectedIdx < len(g.foundRooms) {
				room := g.foundRooms[g.selectedIdx]
				conn, err := JoinRoom(room)
				if err != nil {
					g.lanState = "failed"
					return nil
				}
				g.conn = conn
				g.role = "client"
				g.Reset(HumanVsLAN)
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StateModeSelect
			g.conn = nil
			g.role = ""
			g.lanState = ""
			g.foundRooms = nil
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
	leftMargin := 40

	scale := 0.7

	drawScaledText := func(s string, x, y int, clr color.Color) {
		bounds := text.BoundString(utils.MplusFont, s)
		img := ebiten.NewImage(bounds.Dx(), bounds.Dy())
		img.Fill(color.Transparent)
		text.Draw(img, s, utils.MplusFont, 0, bounds.Dy(), clr)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(x), float64(y)-float64(bounds.Dy())*scale)
		screen.DrawImage(img, op)
	}

	switch g.lanState {
	case "hosting":
		drawScaledText("Hosting... Waiting for player to join.", leftMargin, y, color.White)
	case "searching":
		drawScaledText("Searching for available rooms...", leftMargin, y, color.White)
	case "ready":
		drawScaledText("Available Rooms (Right-click to refresh):", leftMargin, y, color.White)
		y += int(30 * scale)
		for i, room := range g.foundRooms {
			ipStr := fmt.Sprintf("Room %d - %s:%d", i+1, room.IP, room.Port)
			var col color.Color = color.White
			if i == g.selectedIdx {
				col = color.RGBA{200, 255, 200, 255}
			}
			drawScaledText(ipStr, leftMargin+20, y, col)
			y += int(25 * scale)
		}

	case "failed":
		drawScaledText("Connection failed. Check network or host.", leftMargin, y, color.RGBA{255, 100, 100, 255})
	default:
		drawScaledText("Press [H] to HOST a game", leftMargin, y, color.White)
		y += int(30 * scale)
		drawScaledText("Right-click to SEARCH rooms", leftMargin, y, color.White)
	}

	drawScaledText("ESC: Back to menu", leftMargin, WindowHeight-40, color.Gray{150})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}
