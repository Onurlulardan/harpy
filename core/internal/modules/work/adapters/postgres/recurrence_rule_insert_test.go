package postgres

import (
	"strings"
	"testing"
)

func TestRecurrenceRuleInsertSQLIsStatic(t *testing.T) {
	if !strings.Contains(insertRecurrenceRuleSQL, "work_recurrence_rules") {
		t.Fatal("missing recurrence SQL")
	}
}
