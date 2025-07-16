// 文件：src/MCT.go
package src

import (
	"math"
	"math/rand"
	"time"
)

type Board [BoardSize][BoardSize]Stone

func BestMove(board Board, forPlayer Stone, difficulty DifficultyLevel) (int, int) {
	if win, _ := checkWin(board); win != 0 {
		return -1, -1
	}
	root := newNode(board, forPlayer)

	// --- CHANGE THE DEADLINE LOGIC ---
	var thinkingTime time.Duration
	switch difficulty {
	case Easy:
		thinkingTime = 1 * time.Second
	case Medium:
		thinkingTime = 3 * time.Second
	default:
		thinkingTime = 1 * time.Second // Default fallback
	}
	deadline := time.Now().Add(thinkingTime)
	// --- END CHANGE ---

	for time.Now().Before(deadline) {
		leaf := selectNode(root)
		winner := simulate(leaf)
		backpropagate(leaf, winner)
	}

	var best *node
	// In MCTS, the most visited node is the most robust choice.
	maxVisits := -1.0
	for _, c := range root.children {
		if c.visits > maxVisits {
			maxVisits = c.visits
			best = c
		}
	}
	if best == nil {
		// This can happen if no simulations are run or the game is over.
		return randomMove(board)
	}
	return best.move[0], best.move[1]
}

type node struct {
	board    Board
	move     [2]int
	player   Stone
	wins     float64
	visits   float64
	children []*node
	parent   *node
	untried  [][2]int
}

func newNode(b Board, p Stone) *node {
	return &node{
		board:   b,
		player:  p,
		untried: legalMoves(b),
	}
}

func selectNode(n *node) *node {
	for len(n.untried) == 0 && len(n.children) > 0 {
		n = bestUCTChild(n)
	}
	if len(n.untried) > 0 {
		return expand(n)
	}
	return n
}

func expand(n *node) *node {
	m := n.untried[0]
	n.untried = n.untried[1:]

	newBoard := n.board
	newBoard[m[0]][m[1]] = n.player

	child := &node{
		board:   newBoard,
		move:    m,
		player:  3 - n.player,
		parent:  n,
		untried: legalMoves(newBoard),
	}
	n.children = append(n.children, child)
	return child
}

func simulate(n *node) Stone {
	b := n.board
	p := n.player
	for {
		if win, ok := checkWin(b); ok {
			return win
		}
		moves := legalMoves(b)
		if len(moves) == 0 {
			return 0
		}
		m := moves[rand.Intn(len(moves))]
		b[m[0]][m[1]] = p
		p = 3 - p
	}
}

func backpropagate(n *node, winner Stone) {
	for n != nil {
		n.visits++
		if winner == 0 {
			n.wins += 0.5
		} else if winner == n.player {
			n.wins += 0
		} else {
			n.wins += 1
		}
		n = n.parent
	}
}

func bestUCTChild(n *node) *node {
	logN := math.Log(n.visits)
	best := -1.0
	var bestN *node
	for _, c := range n.children {
		uct := c.wins/c.visits + 1.41*math.Sqrt(logN/c.visits)
		if uct > best {
			best = uct
			bestN = c
		}
	}
	return bestN
}

func legalMoves(b Board) [][2]int {
	var moves [][2]int
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] == Empty {
				moves = append(moves, [2]int{r, c})
			}
		}
	}
	rand.Shuffle(len(moves), func(i, j int) { moves[i], moves[j] = moves[j], moves[i] })
	return moves
}

func checkWin(b Board) (Stone, bool) {
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] == Empty {
				continue
			}
			p := b[r][c]
			for _, d := range dirs {
				cnt := 1
				for step := 1; step < 5; step++ {
					nr, nc := r+d[0]*step, c+d[1]*step
					if nr < 0 || nr >= BoardSize || nc < 0 || nc >= BoardSize || b[nr][nc] != p {
						break
					}
					cnt++
				}
				if cnt >= 5 {
					return p, true
				}
			}
		}
	}
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] == Empty {
				return 0, false
			}
		}
	}
	return 0, true
}

func randomMove(b Board) (int, int) {
	moves := legalMoves(b)
	if len(moves) == 0 {
		return -1, -1
	}
	m := moves[rand.Intn(len(moves))]
	return m[0], m[1]
}
