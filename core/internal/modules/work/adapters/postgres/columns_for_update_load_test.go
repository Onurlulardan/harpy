package postgres

import (
	"strings"
	"testing"
)

func TestColumnsForUpdateLoadSQLLocksRows(t *testing.T) {
	if !strings.Contains(listColumnsForUpdateSQL, "FOR UPDATE") {
		t.Fatal("column rows are not locked")
	}
}
