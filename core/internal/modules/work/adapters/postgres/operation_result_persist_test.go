package postgres

import (
	"strings"
	"testing"
)

func TestOperationResultPersistSQLIsStatic(t *testing.T) {
	if !strings.Contains(completeOperationSQL, "work_operations") {
		t.Fatal("missing work operation SQL")
	}
}
