package restoreexecution

import (
	"strings"
	"testing"
)

func TestImportReceiptUsesRestoredDatabaseRoutine(t *testing.T) {
	if !strings.Contains(importReceiptSQL, "platformops_import_restore_receipt_v3") ||
		!strings.Contains(importReceiptSQL, "$17::jsonb") {
		t.Fatal("receipt import is not component-bound")
	}
}
