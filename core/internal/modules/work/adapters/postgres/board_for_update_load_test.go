package postgres

import (
	"strings"
	"testing"
)

func TestBoardForUpdateLoadSQLLocksRow(t *testing.T) {
	if !strings.Contains(getBoardForUpdateSQL, "FOR UPDATE") {
		t.Fatal("board row is not locked")
	}
}
