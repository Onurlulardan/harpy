package systemtransfer

import (
	"testing"
	"time"
)

func TestExportEffectReceiptIdentityMustDifferFromAttempt(t *testing.T) {
	now := time.Date(2026, time.July, 18, 9, 0, 0, 123000000, time.UTC)
	worker := &ExportWorker{clock: exportClockFixture{now: now},
		identities: exportIdentityFixture{identity: fixtureAttemptID}}
	identity, err := worker.newExportEffectReceiptID(fixtureAttemptID)
	if err == nil || identity != [16]byte{} {
		t.Fatalf("identity=%s err=%v", identity, err)
	}
}
