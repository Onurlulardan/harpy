package systemtransfer

import (
	"strings"
	"testing"
)

func TestPersistOutboundFailurePinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(outboundFinishSQL,
		"iam_system_record_export_worker_finish_v1") {
		t.Fatal("export finish does not pin the owned routine")
	}
}
