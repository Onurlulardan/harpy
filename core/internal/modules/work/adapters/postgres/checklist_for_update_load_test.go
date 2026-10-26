package postgres

import (
	"strings"
	"testing"
)

func TestChecklistForUpdateLoadSQLLocksRows(t *testing.T) {
	if !strings.Contains(listChecklistForUpdateSQL, "FOR UPDATE") {
		t.Fatal("checklist rows are not locked")
	}
}
