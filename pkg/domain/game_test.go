package domain

import (
	"testing"
)

func assertLogError(t *testing.T, name, expected, got any) {
	t.Helper()
	if got != expected {
		t.Fatalf("expected %s , %s got %s", name, expected, got)
	}
}

func assertEmptyBoardCheck(t *testing.T, Board [][]CellState) {
	t.Helper()

	if Board[0][0] != Empty {
		t.Fatalf("expected Board empty but got %v", Board[0][0])
	}

}

func TestNewGame(t *testing.T) {
	player1 := "A"
	player2 := "B"
	roomId := "123"
	game := NewGame(player1, player2, roomId)

	assertLogError(t, "roomId", roomId, game.ID)
	assertLogError(t, "Active player", player1, game.ActivePlayer)
	assertLogError(t, "winner", "", game.Winner)
	assertLogError(t, "Boards Size", 2, len(game.Boards))
	assertEmptyBoardCheck(t, game.Boards[player1])
	assertEmptyBoardCheck(t, game.Boards[player2])
}

