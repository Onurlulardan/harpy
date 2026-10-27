package postgres

import (
	"strings"
	"testing"
)

func TestReadLegacyBoardDetailIsWorkOwnedAndRelational(t *testing.T) {
	for _, fragment := range []string{
		"'board'", "'columns'", "'tasks'", "work_task_labels",
		"work_checklist_items", "work_comments", "work_board_memberships",
		"work_task_compatibility_metadata", "compatibility_form_filter",
		"compatibility_is_recurring", "compatibility_recurrence_rule",
	} {
		if !strings.Contains(readVisibleBoardDetailSQL, fragment) {
			t.Fatalf("board detail query missing %q", fragment)
		}
	}
	for _, forbidden := range []string{
		"from iam_", "join iam_", "from forms_", "join forms_",
		"from workflow_", "join workflow_",
	} {
		if strings.Contains(strings.ToLower(readVisibleBoardDetailSQL), forbidden) {
			t.Fatalf("board detail directly reads external owner %q", forbidden)
		}
	}
	if strings.Contains(readVisibleBoardDetailSQL, "'form_slug'") ||
		strings.Contains(readVisibleBoardDetailSQL, "'form_data_id'") {
		t.Fatal("board task form aliases must be enriched through the Forms owner")
	}
}
