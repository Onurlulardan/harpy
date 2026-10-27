package systemtransfer

import (
	"strings"
	"testing"
)

func TestExportRequestPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(submitExportSQL, "iam_system_record_export_submit_v1") {
		t.Fatal("submit export SQL does not pin the owned routine")
	}
}
