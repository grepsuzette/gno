package minesweeper

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	Width  = 8
	Height = 8
	Mines  = 10
)

type Cell struct {
	IsMine        bool
	IsRevealed    bool
	AdjacentMines int
}

type Board [][]Cell

func NewBoard() Board {
	board := make(Board, Height)
	for i := range board {
		board[i] = make([]Cell, Width)
	}
	return board
}

func (b Board) PlaceMines() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < Mines; i++ {
		x, y := rand.Intn(Width), rand.Intn(Height)
		if b[y][x].IsMine {
			i-- // Retry if mine is already placed
			continue
		}
		b[y][x].IsMine = true
		b.UpdateAdjacentMines(x, y)
	}
}

func (b Board) UpdateAdjacentMines(x, y int) {
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}
			adjX, adjY := x+j, y+i
			if adjX >= 0 && adjX < Width && adjY >= 0 && adjY < Height {
				b[adjY][adjX].AdjacentMines++
			}
		}
	}
}

func (b Board) Reveal(x, y int) {
	if b[y][x].IsRevealed || b[y][x].IsMine {
		return
	}
	b[y][x].IsRevealed = true
	if b[y][x].AdjacentMines == 0 {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				if i == 0 && j == 0 {
					continue
				}
				adjX, adjY := x+j, y+i
				if adjX >= 0 && adjX < Width && adjY >= 0 && adjY < Height {
					b.Reveal(adjX, adjY)
				}
			}
		}
	}
}

func (b Board) Print() {
	for y := range b {
		for x := range b[y] {
			if b[y][x].IsRevealed {
				if b[y][x].IsMine {
					fmt.Print("* ")
				} else {
					fmt.Print(b[y][x].AdjacentMines, " ")
				}
			} else {
				fmt.Print(". ")
			}
		}
		fmt.Println("")
	}
}

func main() {
	board := NewBoard()
	board.PlaceMines()

	var x, y int
	for {
		board.Print()
		fmt.Print("Enter coordinates (x y): ")
		_, err := fmt.Scanf("%d %d", &x, &y)
		if err != nil || x < 0 || x >= Width || y < 0 || y >= Height {
			fmt.Println("Invalid input. Please enter valid coordinates.")
			continue
		}
		board.Reveal(x, y)
		if board[y][x].IsMine {
			fmt.Println("Game Over! You hit a mine.")
			break
		}
	}
	board.Print()
}
