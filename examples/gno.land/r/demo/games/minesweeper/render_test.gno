package minesweeper

import (
	"testing"

	"gno.land/p/demo/urequire"
)

func TestParsePath(t *testing.T) {
	board, x, y, debug := parsePath("")
	urequire.True(t, board != nil)
	urequire.Equal(t, x, -1)
	urequire.Equal(t, y, -1)
}
