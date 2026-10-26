package systemtransfer

import (
	"strings"
	"testing"
)

func TestImportRequestPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(submitImportSQL, "iam_system_record_import_submit_v1") {
		t.Fatal("submit import SQL does not pin the owned routine")
	}
}
