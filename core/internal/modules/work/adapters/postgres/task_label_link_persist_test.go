package postgres

import (
	"strings"
	"testing"
)

func TestTaskLabelLinkPersistSQLIsStatic(t *testing.T) {
	if !strings.Contains(persistTaskLabelLinkSQL, "work_task_labels") ||
		strings.Contains(strings.ToUpper(persistTaskLabelLinkSQL), "DELETE") {
		t.Fatal("missing task label SQL")
	}
}
