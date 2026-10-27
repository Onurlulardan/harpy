package systemtransfer

import (
	"context"
	"testing"
	"time"
)

func TestStartSourceCleanupCompletesExactDeleteBeforeDrain(t *testing.T) {
	target, deleted, now := importCleanupFixture()
	store := &importCleanupStoreStub{completeHit: make(chan struct{}, 1),
		scans: []SourceCleanupScan{{ScannedTenantID: target.TenantID,
			Target: target, Found: true}}}
	objects := &importCleanupObjectStub{result: deleted}
	values := importCleanupValuesStub{now: now, id: target.LeaseOwnerID}
	runner, err := NewSourceCleanupRunner(
		store, objects, values, values, importCleanupJitterStub{},
		openImportCleanupAdmission(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-store.completeHit:
	case <-time.After(time.Second):
		t.Fatal("cleanup completion was not persisted")
	}
	runner.BeginDrain()
	if err := runner.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if objects.key != target.ObjectKey || objects.version != target.ObjectVersion {
		t.Fatalf("deleted identity = %q@%q", objects.key, objects.version)
	}
}

func openImportCleanupAdmission(t *testing.T) *sourceCleanupAdmissionStub {
	t.Helper()
	gate := &sourceCleanupAdmissionStub{}
	if !gate.Open() {
		t.Fatal("IAM import cleanup admission did not open")
	}
	return gate
}
