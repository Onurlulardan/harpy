package systemtransfer

import (
	"strings"
	"testing"
)

func TestCheckpointExportPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(outboundCheckpointSQL,
		"iam_system_record_export_worker_checkpoint_v1") {
		t.Fatal("export checkpoint does not pin the owned routine")
	}
}
