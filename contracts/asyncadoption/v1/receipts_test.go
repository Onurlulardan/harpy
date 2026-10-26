package asyncadoptionv1

import "testing"

func TestCancelDispositionRequiresReasonAndTicket(t *testing.T) {
	reasonCode := "V1_DUPLICATE_OCCURRENCE"
	reason := "duplicate schedule occurrence is cancelled during adoption"
	ticket := "CHG-2026-0715"
	receipt, err := NewDispositionReceipt(DispositionReceipt{
		RunID: testRunID, InventoryClass: ClassQueue, SourceName: V1ScheduleQueue,
		SourceItemID: "schedule-1", ItemSHA256: testDigest,
		IdempotencyIdentity: "schedule/occurrence-1", Disposition: DispositionCancelled,
		ReasonCode: &reasonCode, Reason: &reason, Ticket: &ticket, DisposedAt: testTime,
	})
	if err != nil || receipt.ReceiptSHA256 == "" {
		t.Fatalf("create cancelled disposition: %#v %v", receipt, err)
	}
	receipt.Ticket = nil
	if receipt.Validate() == nil {
		t.Fatal("reason-only cancellation was accepted without a ticket")
	}
}

func TestRestartFenceZeroObservationAndAcceptanceAreDigestBound(t *testing.T) {
	fences := []RestartFenceReceipt{
		newTestFence(t, FenceEnqueue, "workflow.v1.enqueue", "2026-07-15T12:00:01Z"),
		newTestFence(t, FenceScheduler, V1CronScheduler, "2026-07-15T12:00:02Z"),
		newTestFence(t, FenceConsumer, V1ScheduleQueue, "2026-07-15T12:00:03Z"),
	}
	componentRoot, err := ComponentManifestSHA256(fences)
	if err != nil {
		t.Fatal(err)
	}
	zero, err := NewZeroInventoryReceipt(ZeroInventoryReceipt{
		RunID: testRunID, InventorySHA256: testDigest,
		ObservedAt: "2026-07-15T12:00:04Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	acceptance, err := NewAcceptanceReceipt(AcceptanceReceipt{
		RunID: testRunID, InventorySHA256: testDigest,
		ComponentManifestSHA256: componentRoot, ZeroInventorySHA256: zero.ReceiptSHA256,
		AcceptedAt:          "2026-07-15T12:00:05Z",
		DispositionReceipts: []string{testDigest},
	})
	if err != nil || acceptance.AcceptanceSHA256 == "" {
		t.Fatalf("create acceptance: %#v %v", acceptance, err)
	}
	acceptance.ZeroInventorySHA256 = componentRoot
	if acceptance.Validate() == nil {
		t.Fatal("acceptance allowed changed zero-inventory evidence")
	}
	zero.QueueCount = 1
	if zero.Validate() == nil {
		t.Fatal("non-zero queue inventory was accepted")
	}
}

func TestRestartFenceComparesCanonicalTimestampsChronologically(t *testing.T) {
	_, err := NewRestartFenceReceipt(RestartFenceReceipt{
		RunID: testRunID, Kind: FenceConsumer, ComponentName: V1ScheduleQueue,
		ComponentContractVersion: "v1.runtime-fence.v1",
		BinarySHA256:             testDigest, ConfigSHA256: testDigest,
		RestartProbeSHA256: testDigest,
		FencedAt:           "2026-07-15T12:00:00Z",
		RestartVerifiedAt:  "2026-07-15T12:00:00.1Z",
	})
	if err != nil {
		t.Fatalf("chronologically later fractional timestamp rejected: %v", err)
	}
}

func newTestFence(t *testing.T, kind FenceKind, name string, verified string) RestartFenceReceipt {
	t.Helper()
	receipt, err := NewRestartFenceReceipt(RestartFenceReceipt{
		RunID: testRunID, Kind: kind, ComponentName: name,
		ComponentContractVersion: "v1.runtime-fence.v1",
		BinarySHA256:             testDigest, ConfigSHA256: testDigest,
		RestartProbeSHA256: testDigest, FencedAt: testTime, RestartVerifiedAt: verified,
	})
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}
