package systemtransfer

import (
	"strings"
	"testing"
)

func TestListReceiptsPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(readReceiptsSQL, "iam_system_record_import_receipts_v1") {
		t.Fatal("receipt SQL does not pin the owned routine")
	}
}
