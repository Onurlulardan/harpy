package postgres

import (
	"strings"
	"testing"
)

func TestProjectionPersistSQLIsStatic(t *testing.T) {
	if !strings.Contains(upsertInboxProjectionSQL, "work_inbox_items") ||
		!strings.Contains(upsertCalendarProjectionSQL, "work_calendar_items") {
		t.Fatal("missing Work projection persistence SQL")
	}
}
