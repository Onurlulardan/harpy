package postgres

import (
	"strings"
	"testing"
)

func TestProjectionRepairFlagSQLIsStatic(t *testing.T) {
	if !strings.Contains(markProjectionRepairSQL, "work_projection_streams") {
		t.Fatal("missing projection SQL")
	}
}
