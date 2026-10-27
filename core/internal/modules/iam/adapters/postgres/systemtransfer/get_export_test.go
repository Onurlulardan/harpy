package systemtransfer

import (
	"strings"
	"testing"
)

func TestGetExportPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(readExportSQL, "iam_system_record_export_read_v1") {
		t.Fatal("read export SQL does not pin the owned routine")
	}
}
