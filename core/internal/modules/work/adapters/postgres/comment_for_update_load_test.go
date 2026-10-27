package postgres

import (
	"strings"
	"testing"
)

func TestCommentForUpdateLoadSQLLocksRow(t *testing.T) {
	if !strings.Contains(getCommentForUpdateSQL, "FOR UPDATE") {
		t.Fatal("comment row is not locked")
	}
}
