package postgres

import (
	"strings"
	"testing"
)

func TestCalendarEventPersistSQLUsesVersionCAS(t *testing.T) {
	if !strings.Contains(updateCalendarEventSQL, "row_version = $15") {
		t.Fatal("missing calendar CAS")
	}
}
