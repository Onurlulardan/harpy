package systemtransfer

import (
	"strings"
	"testing"
)

func TestPersistOutcomeUnknownPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(outboundUnknownSQL,
		"iam_system_record_export_worker_unknown_v1") {
		t.Fatal("export unknown outcome does not pin the owned routine")
	}
}
