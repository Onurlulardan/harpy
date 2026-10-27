package systemtransfer

import (
	"context"
	"testing"
	"time"
)

func TestSourceCleanupDoesNotScanUntilAdmissionOpens(t *testing.T) {
	target, deleted, now := importCleanupFixture()
	store := &importCleanupStoreStub{claimHit: make(chan struct{}, 1),
		completeHit: make(chan struct{}, 1), scans: []SourceCleanupScan{{
			ScannedTenantID: target.TenantID, Target: target, Found: true,
		}}}
	gate := &sourceCleanupAdmissionStub{}
	values := importCleanupValuesStub{now: now, id: target.LeaseOwnerID}
	runner, err := NewSourceCleanupRunner(store,
		&importCleanupObjectStub{result: deleted}, values, values,
		importCleanupJitterStub{}, gate)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-store.claimHit:
		t.Fatal("closed admission scanned IAM import cleanup store")
	case <-time.After(2 * sourceCleanupAdmissionPoll):
	}
	if !gate.Open() {
		t.Fatal("IAM import cleanup admission did not open")
	}
	select {
	case <-store.completeHit:
	case <-time.After(time.Second):
		t.Fatal("opened admission did not resume IAM import cleanup")
	}
	runner.BeginDrain()
	if err := runner.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSourceCleanupRequiresAdmissionGate(t *testing.T) {
	target, _, now := importCleanupFixture()
	values := importCleanupValuesStub{now: now, id: target.LeaseOwnerID}
	_, err := NewSourceCleanupRunner(&importCleanupStoreStub{},
		&importCleanupObjectStub{}, values, values, importCleanupJitterStub{}, nil)
	if err == nil {
		t.Fatal("nil IAM import cleanup admission gate was accepted")
	}
}
