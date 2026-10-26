package restoreexecution

import (
	"strings"
	"testing"
)

func TestCompleteUsesSignedReceiptRoutine(t *testing.T) {
	if !strings.Contains(completeSQL, "platformops_complete_restore_execution_v2") ||
		!strings.Contains(completeSQL, "$5") {
		t.Fatal("completion does not bind signed receipt bytes")
	}
}
