package postgres

import (
	"strings"
	"testing"
)

func TestTaskPersistSQLUsesVersionCAS(t *testing.T) {
	if !strings.Contains(updateTaskSQL, "row_version = $12") {
		t.Fatal("missing task CAS")
	}
}

func TestTaskPersistenceCarriesCreatorPriorityAndStart(t *testing.T) {
	for _, fragment := range []string{"creator_user_id", "priority", "starts_at"} {
		if !strings.Contains(insertTaskSQL, fragment) || !strings.Contains(updateTaskSQL, fragment) &&
			fragment != "creator_user_id" {
			t.Fatalf("task persistence missing %q", fragment)
		}
	}
	if !strings.Contains(getTaskForUpdateSQL,
		"creator_user_id, task_kind, status_category, priority") {
		t.Fatal("task lock load omits typed compatibility fields")
	}
}
