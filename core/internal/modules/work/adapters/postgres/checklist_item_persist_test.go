package postgres

import (
	"strings"
	"testing"
)

func TestChecklistItemPersistSQLUsesVersionCAS(t *testing.T) {
	if !strings.Contains(updateChecklistItemSQL, "row_version = $7") {
		t.Fatal("missing checklist CAS")
	}
}
