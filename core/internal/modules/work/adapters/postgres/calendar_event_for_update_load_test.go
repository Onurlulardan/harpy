package postgres

import (
	"strings"
	"testing"
)

func TestCalendarEventForUpdateLoadSQLLocksRow(t *testing.T) {
	if !strings.Contains(getCalendarEventForUpdateSQL, "FOR UPDATE") {
		t.Fatal("calendar row is not locked")
	}
}
