package postgres

import (
	"strings"
	"testing"
)

func TestChecklistArchivePersistSQLIsStatic(t *testing.T) {
	if !strings.Contains(persistChecklistArchiveSQL, "work_checklist_items") {
		t.Fatal("missing checklist SQL")
	}
}
