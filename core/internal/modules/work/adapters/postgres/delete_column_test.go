package postgres

import (
	"strings"
	"testing"
)

func TestDeleteColumnRequiresVersionAndNoTasks(t *testing.T) {
	for _, fragment := range []string{
		"column_row.row_version = $2", "NOT EXISTS", "work_tasks",
	} {
		if !strings.Contains(deleteEmptyColumnSQL, fragment) {
			t.Fatalf("delete column query missing %q", fragment)
		}
	}
}
