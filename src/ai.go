package src

import (
	// "fmt"
	"math/rand"
)

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
