package src

import (
	"bytes" // ADDED
	_ "embed"
	"fmt"
	"image/color"
	"log" // ADDED
	"math"
	"net"
	"os"
	"time"
	"wuziqi/utils"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)
//go:embed assets/bgm.mp3
var bgmBytes []byte

//go:embed assets/move.wav
var stoneBytes []byte

type Game struct {
	board            [BoardSize][BoardSize]Stone
	currentTurn      Stone
	winner           Stone
	moves            int
	state            GameState
	playMode         PlayMode
	difficulty       DifficultyLevel
	moveHistory      [][2]int
	pendingAI        bool
	conn             net.Conn
	role             string
	lanState         string
	foundRooms       []RoomInfo
	selectedIdx      int
	lanReceivedMoves chan [2]int
	undoRequested    bool
	undoPending      bool
	undoResponseCh   chan bool
	lastMover        Stone
	audioContext *audio.Context
	bgmPlayer    *audio.Player
	stonePlayer  *audio.Player
	masterVolume float64
}

func (g *Game) initAudio() {
	g.audioContext = audio.NewContext(48000)

	// --- BGM ---
	if bgmBytes != nil {
		bgmStream, err := mp3.DecodeWithSampleRate(g.audioContext.SampleRate(), bytes.NewReader(bgmBytes))
		if err != nil {
			log.Printf("audio warning: failed to decode bgm: %v\n", err)
		} else {
			bgmLoop := audio.NewInfiniteLoop(bgmStream, bgmStream.Length())
			g.bgmPlayer, err = g.audioContext.NewPlayer(bgmLoop)
			if err != nil {
				log.Printf("audio warning: failed to create bgm player: %v\n", err)
			} else {
				// Use master volume (BGM is quieter)
				g.bgmPlayer.SetVolume(g.masterVolume * 0.5)
				g.bgmPlayer.Play()
			}
		}
	}

	// --- Stone Sound ---
	if stoneBytes == nil {
		log.Println("audio warning: move.wav file not found. Check that `src/assets/move.wav` exists.")
		return
	}

	stoneStream, err := wav.DecodeWithSampleRate(g.audioContext.SampleRate(), bytes.NewReader(stoneBytes))
	if err != nil {
		log.Printf("audio warning: failed to decode move.wav (is it a valid WAV file?): %v\n", err)
	} else {
		g.stonePlayer, err = g.audioContext.NewPlayer(stoneStream)
		if err != nil {
			log.Printf("audio warning: failed to create stone sfx player: %v\n", err)
		} else {
			// Use master volume
			g.stonePlayer.SetVolume(g.masterVolume)
		}
	}
}

func NewGame() *Game {
	utils.InitFont()
	g := &Game{
		state:            StateModeSelect,
		lanReceivedMoves: make(chan [2]int, 10),
		masterVolume:     0.5, // Default volume is 50%
	}
	g.initAudio()
	return g
}

func (g *Game) Reset(mode PlayMode) {
	fmt.Printf("[RESET] mode=%v role=%v conn=%v\n", mode, g.role, g.conn != nil)
	g.board = [BoardSize][BoardSize]Stone{}
	g.currentTurn = Black
	g.winner = Empty
	g.moves = 0
	g.playMode = mode
	g.state = StatePlaying
	g.moveHistory = nil
	g.pendingAI = false

	if mode == HumanVsLAN && g.conn != nil {
		g.lanReceivedMoves = make(chan [2]int, 10)

		g.undoResponseCh = make(chan bool, 1)

		go func() {
			for {
				if g.conn == nil {
					return
				}
				row, col, op, err := recvMessage(g.conn)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					fmt.Println("recvMessage error:", err)
					g.lanState = "failed"
					return
				}
				switch op {
				case "MOVE":
					g.lanReceivedMoves <- [2]int{row, col}
				case "UNDO_REQUEST":
					g.undoPending = true
					g.undoRequested = false
				case "UNDO_ACCEPT":
					g.undoResponseCh <- true
				case "UNDO_REJECT":
					g.undoResponseCh <- false
				case "PEER_LEFT":
					g.lanState = "peerLeft"
				}
			}
		}()
	}
}

func (g *Game) Update() error {
	switch g.state {
	case StateModeSelect:
		if g.conn != nil {
			_ = g.conn.Close()
		}
		g.conn = nil
		g.role = ""
		g.lanState = ""
		g.foundRooms = nil
		g.board = [BoardSize][BoardSize]Stone{}
		g.currentTurn = Black
		g.winner = Empty
		g.moves = 0
		g.moveHistory = nil
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
				volumeChanged := false
		// Check for volume up keys
		if inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) || inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
			g.masterVolume += 0.1
			volumeChanged = true
		}
		// Check for volume down keys
		if inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) || inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
			g.masterVolume -= 0.1
			volumeChanged = true
		}

		if volumeChanged {
			// Clamp volume between 0.0 and 1.0
			if g.masterVolume > 1.0 {
				g.masterVolume = 1.0
			}
			if g.masterVolume < 0.0 {
				g.masterVolume = 0.0
			}

			// Apply new volume to players
			if g.bgmPlayer != nil {
				g.bgmPlayer.SetVolume(g.masterVolume * 0.5)
			}
			if g.stonePlayer != nil {
				g.stonePlayer.SetVolume(g.masterVolume)
			}

			// Apply new volume to players
			if g.bgmPlayer != nil {
				g.bgmPlayer.SetVolume(g.masterVolume * 0.5)
			}
			if g.stonePlayer != nil {
				g.stonePlayer.SetVolume(g.masterVolume)
			}
		}
		if g.playMode == HumanVsAI && g.currentTurn == White && !g.pendingAI && g.winner == Empty {
			g.pendingAI = true
		}

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
			g.cleanupLAN()
			g.state = StateModeSelect
			return nil
		}

		if g.playMode == HumanVsLAN {
			isMyTurn := (g.role == "host" && g.currentTurn == Black) ||
				(g.role == "client" && g.currentTurn == White)

			if isMyTurn {
				if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
					g.handlePlayerMove()
				}
			} else {
				select {
				case move := <-g.lanReceivedMoves:
					g.board[move[0]][move[1]] = g.currentTurn
					g.moveHistory = append(g.moveHistory, move)
					g.moves++
					if g.checkWin(move[0], move[1]) {
						g.winner = g.currentTurn
						g.state = StateGameOver
					} else if g.moves == BoardSize*BoardSize {
						g.winner = Empty
						g.state = StateGameOver
					} else {
						g.currentTurn = 3 - g.currentTurn
					}
					fmt.Printf("[RECV] %s received: (%d,%d)\n", g.role, move[0], move[1])
				default:
				}
			}
			if g.undoPending && g.undoRequested {
				select {
				case accept := <-g.undoResponseCh:
					g.undoPending = false
					if accept {
						g.undoLastMove()
					}
				default:
					return nil
				}
			}

			if g.undoPending && !g.undoRequested {
				if inpututil.IsKeyJustPressed(ebiten.KeyY) {
					sendUndoAccept(g.conn)
					g.undoLastMove()
					g.undoPending = false
				} else if inpututil.IsKeyJustPressed(ebiten.KeyN) {
					sendUndoReject(g.conn)
					g.undoPending = false
				}
				return nil
			}
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

				go func() {
					for {
						if g.conn == nil {
							break
						}
						row, col, op, err := recvMessage(g.conn)
						if err != nil {
							g.lanState = "failed"
							break
						}
						switch op {
						case "MOVE":
							g.lanReceivedMoves <- [2]int{row, col}
						case "UNDO_REQUEST":
							g.undoPending = true
							g.undoRequested = false
						case "UNDO_ACCEPT":
							g.undoResponseCh <- true
						case "UNDO_REJECT":
							g.undoResponseCh <- false
						}
					}
				}()
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

				go func() {
					for {
						if g.conn == nil {
							break
						}
						row, col, op, err := recvMessage(g.conn)
						if err != nil {
							g.lanState = "failed"
							break
						}
						switch op {
						case "MOVE":
							g.lanReceivedMoves <- [2]int{row, col}
						case "UNDO_REQUEST":
							g.undoPending = true
							g.undoRequested = false
						case "UNDO_ACCEPT":
							g.undoResponseCh <- true
						case "UNDO_REJECT":
							g.undoResponseCh <- false
						}
					}
				}()
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.cleanupLAN()
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
	if g.undoPending {
		return
	}
	x, y := ebiten.CursorPosition()
	col := int(math.Round((float64(x) - Margin) / TileSize))
	row := int(math.Round((float64(y) - Margin) / TileSize))
	if row >= 0 && row < BoardSize && col >= 0 && col < BoardSize && g.board[row][col] == Empty {
		g.placeStoneAt(row, col)
	}
}

func (g *Game) placeStoneAt(row, col int) {
	if g.board[row][col] != Empty {
		return
	}

	isMyTurn := (g.role == "host" && g.currentTurn == Black) ||
		(g.role == "client" && g.currentTurn == White)

	if g.playMode == HumanVsLAN && !isMyTurn {
		return
	}

	g.board[row][col] = g.currentTurn
	g.moveHistory = append(g.moveHistory, [2]int{row, col})
	g.moves++

	// --- Play sound effect ---
	if g.stonePlayer != nil {
		log.Println("Attempting to play move sound...") // ADD THIS LINE FOR DEBUGGING
		g.stonePlayer.Rewind()
		g.stonePlayer.Play()
	}
	// --- End of sound effect ---

	if g.playMode == HumanVsLAN && g.conn != nil {
		sendMove(g.conn, row, col)
		fmt.Printf("[SEND] %s sent: (%d,%d)\n", g.role, row, col)
	}
	g.lastMover = g.currentTurn
	if g.checkWin(row, col) {
		g.winner = g.currentTurn
		g.state = StateGameOver
	} else if g.moves == BoardSize*BoardSize {
		g.winner = Empty
		g.state = StateGameOver
	} else {
		g.currentTurn = 3 - g.currentTurn
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
		if g.undoPending {
			ebitenutil.DrawRect(screen, 0, 0,
				float64(WindowWidth), float64(WindowHeight),
				color.RGBA{0, 0, 0, 180})

			drawSmallCenter := func(lines []string) {
				scale := 0.5
				lineH := int(float64(utils.MplusFont.Metrics().Height>>6) * scale)
				totalH := lineH * len(lines)
				startY := (WindowHeight - totalH) / 2

				for i, txt := range lines {
					b := text.BoundString(utils.MplusFont, txt)
					img := ebiten.NewImage(b.Dx(), b.Dy())
					img.Fill(color.Transparent)
					text.Draw(img, txt, utils.MplusFont, 0, b.Dy(), color.White)

					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(
						float64(WindowWidth/2)-float64(b.Dx())*scale/2,
						float64(startY+i*lineH),
					)
					screen.DrawImage(img, op)
				}
			}
			if g.playMode == HumanVsLAN && g.lanState == "peerLeft" {
				ebitenutil.DrawRect(screen, 0, 0,
					float64(WindowWidth), float64(WindowHeight),
					color.RGBA{0, 0, 0, 180})
				drawSmallCenter([]string{
					"Opponent has left the game",
					"Click anywhere to return to menu",
				})
			}
			if g.undoRequested {
				drawSmallCenter([]string{"Waiting for opponent to accept undo..."})
			} else {
				drawSmallCenter([]string{
					"Opponent wants to undo last move",
					"Press [Y] to accept  |  [N] to reject",
				})
			}
		}
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
func (g *Game) whoAmI() Stone {
	if g.role == "host" {
		return Black
	}
	return White
}
func (g *Game) undoMove() {
	if g.state != StatePlaying || len(g.moveHistory) == 0 {
		return
	}

	switch g.playMode {
	case HumanVsLAN:
		isMyTurn := (g.role == "host" && g.currentTurn == Black) ||
			(g.role == "client" && g.currentTurn == White)
		if !isMyTurn && !g.undoPending && !g.undoRequested &&
			g.lastMover == g.whoAmI() {
			g.undoRequested = true
			g.undoPending = true
			_ = sendUndoRequest(g.conn)
		}
		return

	default:
		steps := 1
		if g.playMode == HumanVsAI && len(g.moveHistory) >= 2 {
			steps = 2
		}
		for i := 0; i < steps; i++ {
			if len(g.moveHistory) == 0 {
				break
			}
			last := g.moveHistory[len(g.moveHistory)-1]
			g.board[last[0]][last[1]] = Empty
			g.moveHistory = g.moveHistory[:len(g.moveHistory)-1]
			g.moves--
			g.currentTurn = 3 - g.currentTurn
		}
	}
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	scale := 0.7
	margin := 20.0
	statusTexts := []string{
		fmt.Sprintf("Volume: %d%% (+/-)", int(g.masterVolume*100)),
		"ESC: Menu",
	}


	lineHeight := text.BoundString(utils.MplusFont, "A").Dy()
	scaledLineHeight := float64(lineHeight) * scale * 1.2

	op := &ebiten.DrawImageOptions{}
	for i, s := range statusTexts {
		bounds := text.BoundString(utils.MplusFont, s)

		x := float64(WindowWidth) - (float64(bounds.Dx()) * scale) - margin
		y := margin + (float64(i) * scaledLineHeight)

		op.GeoM.Reset()
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)


		op.ColorM.Reset()
		op.ColorM.ScaleWithColor(color.Black)

		text.DrawWithOptions(screen, s, utils.MplusFont, op)
	}


	turnText := "Current Turn: "
	cx := float64(text.BoundString(utils.MplusFont, turnText).Dx() + 40)
	cy := float64(WindowWidth + StatusHeight/2)
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

func (g *Game) undoLastMove() {
	if len(g.moveHistory) == 0 {
		return
	}
	last := g.moveHistory[len(g.moveHistory)-1]
	g.board[last[0]][last[1]] = Empty
	g.moveHistory = g.moveHistory[:len(g.moveHistory)-1]
	g.moves--
	g.currentTurn = 3 - g.currentTurn
	g.lastMover = g.currentTurn
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
		drawScaledText("Connection failed", leftMargin, y, color.RGBA{255, 100, 100, 255})
	default:
		drawScaledText("Press [H] to HOST a game", leftMargin, y, color.White)
		y += int(30 * scale)
		drawScaledText("Right-click to SEARCH rooms", leftMargin, y, color.White)
	}

	drawScaledText("ESC: Back to menu", leftMargin, WindowHeight-40, color.Gray{150})
}

func (g *Game) cleanupLAN() {
	if g.conn != nil {
		_ = g.conn.Close()
		g.conn = nil
	}
	g.role = ""
	g.lanState = ""
	g.foundRooms = nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}
