package postgres

import (
	"strings"
	"testing"
)

func TestTaskForUpdateLoadSQLLocksRow(t *testing.T) {
	if !strings.Contains(getTaskForUpdateSQL, "FOR UPDATE") {
		t.Fatal("task row is not locked")
	}
}
