package postgres

import (
	"strings"
	"testing"
)

func TestLabelArchivePersistSQLIsStatic(t *testing.T) {
	if !strings.Contains(persistLabelArchiveSQL, "work_labels") {
		t.Fatal("missing label SQL")
	}
}
