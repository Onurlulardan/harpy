package systemtransfer

import (
	"strings"
	"testing"
)

func TestPersistOutboundSuccessPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(outboundCompleteSQL,
		"iam_system_record_export_worker_complete_v1") {
		t.Fatal("export completion does not pin the owned routine")
	}
}
