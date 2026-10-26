package recovery

import (
	"strings"
	"testing"
)

func TestRecordCopyStatementUsesRecoveryReceiptRoutine(t *testing.T) {
	if !strings.Contains(recordCopySQL, "documents_recovery_record_copy_v1") {
		t.Fatal("copy receipt routine is not bound")
	}
}
