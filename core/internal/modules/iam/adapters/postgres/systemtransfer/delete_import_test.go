package systemtransfer

import (
	"strings"
	"testing"
)

func TestDeleteImportPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(cancelImportSQL, "iam_system_record_import_cancel_v1") {
		t.Fatal("cancel import SQL does not pin the owned routine")
	}
}
