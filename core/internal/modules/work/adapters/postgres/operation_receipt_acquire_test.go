package postgres

import (
	"strings"
	"testing"
)

func TestOperationReceiptAcquireSQLIsStatic(t *testing.T) {
	if !strings.Contains(claimOperationSQL, "work_operations") {
		t.Fatal("missing work operation SQL")
	}
}
