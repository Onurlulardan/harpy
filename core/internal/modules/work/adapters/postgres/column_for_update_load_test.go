package postgres

import (
	"strings"
	"testing"
)

func TestColumnForUpdateLoadSQLLocksRow(t *testing.T) {
	if !strings.Contains(getColumnForUpdateSQL, "FOR UPDATE") {
		t.Fatal("column row is not locked")
	}
}
