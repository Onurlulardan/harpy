package systemtransfer

import (
	"strings"
	"testing"
)

func TestAcquireOutboundPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(outboundClaimSQL,
		"iam_system_record_export_worker_claim_v1") {
		t.Fatal("export claim does not pin the owned routine")
	}
}
