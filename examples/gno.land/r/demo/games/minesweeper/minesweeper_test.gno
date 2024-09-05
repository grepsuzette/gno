package minesweeper

import (
	"testing"

	"gno.land/p/demo/urequire"
)

func TestBoard(t *testing.T) {
	{
		board := NewBoard()
		urequire.False(t, board == nil)
		urequire.Equal(t, len(*board), Height)
		urequire.Equal(t, len(board[0]), Width)
	}
	{
		var board Board
		urequire.NotPanics(t, func() {
			board = BoardFromRepr("")
		})
		// 	// check all squares unrevealed
		// 	// check some mines were placed
		// 	foundMines := false
		// 	for i := 0; i < Width*Height; i++ {
		// 		urequire.False(t, board[i/3][i%3].IsRevealed)
		// 		if board[i/3][i%3].IsMine {
		// 			foundMines = true
		// 		}
		// 	}
		// 	urequire.False(t, foundMines)

		// 	// check BoardFromRepr(hex).ToHexRepr() == hex
		// 	hex := board.ToHexRepr()
		// 	urequire.Equal(t, len(hex), 2*Width*Height)
		// 	urequire.Equal(t, hex, BoardFromRepr(hex).ToHexRepr())
		// }
		// {
		// 	urequire.PanicsWithMessage(t, "invalid state length", func() {
		// 		BoardFromRepr("xx")
		// 	})
	}
}
