package systemtransfer

import (
	"strings"
	"testing"
)

func TestResolveImportPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(readImportSQL, "iam_system_record_import_read_v1") {
		t.Fatal("read import SQL does not pin the owned routine")
	}
}
