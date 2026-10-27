package systemtransfer

import (
	"strings"
	"testing"
)

func TestCancelExportPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(cancelExportSQL, "iam_system_record_export_cancel_v1") {
		t.Fatal("cancel export SQL does not pin the owned routine")
	}
}
