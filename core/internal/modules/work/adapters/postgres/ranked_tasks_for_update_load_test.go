package postgres

import (
	"strings"
	"testing"
)

func TestRankedTasksForUpdateLoadSQLLocksRows(t *testing.T) {
	if !strings.Contains(listRankedTasksForUpdateSQL, "FOR UPDATE") {
		t.Fatal("task rows are not locked")
	}
}
