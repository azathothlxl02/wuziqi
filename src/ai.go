package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
)

func GetAIMove(board [BoardSize][BoardSize]Stone) (row, col int) {
	flat := make([]int, BoardSize*BoardSize)
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			flat[r*BoardSize+c] = int(board[r][c])
		}
	}

	payload, _ := json.Marshal(map[string]interface{}{"board": flat})

	cmd := exec.Command("python", "src/go_call_np.py")
	cmd.Stdin = bytes.NewReader(payload)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Python error:", string(out), err)
		return -1, -1
	}

	var resp struct{ Row, Col int }
	json.Unmarshal(out, &resp)
	// fmt.Println("AI move:", resp.Row, resp.Col)
	return resp.Row, resp.Col
}
func GetRandomAIMove(board [BoardSize][BoardSize]Stone) (int, int) {
	empty := make([][2]int, 0)
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if board[r][c] == Empty {
				empty = append(empty, [2]int{r, c})
			}
		}
	}
	if len(empty) == 0 {
		return -1, -1
	}
	move := empty[rand.Intn(len(empty))]
	// fmt.Println("AI move:", move[0], move[1])
	return move[0], move[1]
	//竖 横
}
