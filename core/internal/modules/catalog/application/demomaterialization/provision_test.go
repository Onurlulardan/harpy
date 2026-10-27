package demomaterialization

import (
	"testing"
	"time"
)

func TestProvisionReceiptPreservesDurableWaitAndReadyStates(t *testing.T) {
	pending, verified, requestSHA := fixtureSnapshot(t, fixtureRequest(t), false)
	if !snapshotMatches(pending, verified, requestSHA) {
		t.Fatal("durable pending snapshot should match")
	}
	pendingReceipt := receiptFromSnapshot(pending)
	if pendingReceipt.Ready || pendingReceipt.RetryAt != pending.RetryAt ||
		pendingReceipt.CompletedAt != pending.CompletedAt {
		t.Fatalf("pending receipt=%+v", pendingReceipt)
	}
	ready := pending
	ready.Ready = true
	ready.RetryAt = time.Time{}
	if !snapshotMatches(ready, verified, requestSHA) {
		t.Fatal("durable ready snapshot should match")
	}
	readyReceipt := receiptFromSnapshot(ready)
	if !readyReceipt.Ready || !readyReceipt.RetryAt.IsZero() {
		t.Fatalf("ready receipt=%+v", readyReceipt)
	}
}
