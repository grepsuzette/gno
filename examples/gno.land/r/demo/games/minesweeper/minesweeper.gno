package minesweeper

import (
	"encoding/hex"
	"net/url"
	"strings"
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

type Board [][]*Cell

func NewBoard() *Board {
	board := make(Board, Height)
	for i := range board {
		board[i] = make([]*Cell, Width)
	}
	return &board
}

// first let's just use plain board
// there are 3 layers.
// input is a hex string where each hex-decoded byte 8 bits is 0b00ZYXXXX
//
//	XXXX: count of adjacent mines (4bits, valid ∈ 0-8),
//	Y: discovered bit
//	Z: isMine bit
func BoardFromRepr(hexstate string) *Board {
	state := stateToBytes(hexstate)
	board := NewBoard()
	if len(state) == 0 {
		board.placeMines()
	} else if len(state) != Width*Height {
		panic("invalid state length")
	} else {
		for i, o := range state {
			board[i/3][i%3] = &Cell{
				IsMine:        ((o & 0b0100000) >> 5) == 1,
				IsRevealed:    ((o & 0b010000) >> 4) == 1,
				AdjacentMines: int(o & 0b001111),
			}
		}
	}
	return board
}

// ToHexRepr is the opposite operation of BoardFromRepr
func (b Board) ToHexRepr() string {
	var builder strings.Builder
	for i := 0; i < Width*Height; i++ {
		cell := b[i/3][i%3]
		var isMine uint8 = 0
		var isRevealed uint8 = 0
		if cell.IsMine {
			isMine = 1
		}
		if cell.IsRevealed {
			isRevealed = 1
		}
		builder.WriteByte(
			0b00000000 | (isMine << 5) | (isRevealed << 4) | (uint8(cell.AdjacentMines)),
		)
	}
	return builder.String()
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

func (b Board) placeMines() {
	for i := 0; i < Mines; i++ {
		x, y := Intn(Width), Intn(Height)
		if b[y][x].IsMine {
			i-- // Retry if mine is already placed
			continue
		}
		b[y][x].IsMine = true
		b.updateAdjacentMines(x, y)
	}
}

func (b Board) updateAdjacentMines(x, y int) {
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

func stateToBytes(urlEncodedHex string) []byte {
	decoded, err := url.QueryUnescape(urlEncodedHex)
	if err != nil {
		panic("Error decoding URL:" + err.Error())
	}
	bytes, err := hex.DecodeString(decoded)
	if err != nil {
		panic("Error decoding hex:" + err.Error())
	}
	return bytes
}
