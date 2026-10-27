package tenantmigrationexecutor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/google/uuid"
)

func TestProcessOneCompletesExactClaim(t *testing.T) {
	storage := &fakeStore{claim: validClaim(), found: true}
	executor := testExecutor(t, storage, fakeMigrator{})
	gate := &admission.Gate{}
	gate.Open()
	permit, _ := gate.Admit()
	processed, err := executor.processOne(
		context.Background(), permit, storage.claim.ProfileInstanceID,
	)
	if err != nil || !processed || storage.completed.JobID != storage.claim.JobID {
		t.Fatalf("process claim: processed=%t completed=%s err=%v",
			processed, storage.completed.JobID, err)
	}
}

func TestProcessOnePersistsAmbiguousOutcome(t *testing.T) {
	storage := &fakeStore{claim: validClaim(), found: true}
	executor := testExecutor(t, storage, fakeMigrator{err: ErrAmbiguousOutcome})
	gate := &admission.Gate{}
	gate.Open()
	permit, _ := gate.Admit()
	processed, err := executor.processOne(
		context.Background(), permit, storage.claim.ProfileInstanceID,
	)
	if err != nil || !processed || storage.failure.Outcome != "reconcile_required" ||
		storage.failure.Code != "migration_outcome_ambiguous" {
		t.Fatalf("ambiguous result: %+v processed=%t err=%v",
			storage.failure, processed, err)
	}
}

func TestRunRetiresReceiptAtDrain(t *testing.T) {
	storage := &fakeStore{registered: make(chan uuid.UUID, 1)}
	executor := testExecutor(t, storage, fakeMigrator{})
	gate := &admission.Gate{}
	gate.Open()
	done := make(chan error, 1)
	go func() { done <- executor.Run(context.Background(), gate) }()
	select {
	case <-storage.registered:
	case <-time.After(time.Second):
		t.Fatal("executor did not register readiness")
	}
	gate.BeginDrain()
	select {
	case err := <-done:
		if err != nil || storage.retired == uuid.Nil {
			t.Fatalf("drain result retired=%s err=%v", storage.retired, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("executor did not drain")
	}
}

type fakeMigrator struct{ err error }

func (migrator fakeMigrator) Migrate(
	context.Context, Claim,
) (MigrationReceipt, error) {
	return MigrationReceipt{ReceiptSHA256: strings.Repeat("a", 64),
		EvidenceSHA256: strings.Repeat("b", 64)}, migrator.err
}

type fakeStore struct {
	claim      Claim
	found      bool
	completed  Claim
	failure    Failure
	registered chan uuid.UUID
	retired    uuid.UUID
}

func (storage *fakeStore) RegisterReady(
	_ context.Context, profileID uuid.UUID, release Release, ttl time.Duration,
) (ReadyReceipt, error) {
	if storage.registered != nil {
		storage.registered <- profileID
	}
	return ReadyReceipt{ProfileInstanceID: profileID, ReleaseID: release.ID,
		ReadyUntil: time.Now().Add(ttl)}, nil
}

func (storage *fakeStore) HeartbeatReady(
	context.Context, uuid.UUID, time.Duration,
) (time.Time, error) {
	return time.Now().Add(ReadyTTL), nil
}

func (storage *fakeStore) RetireReady(
	_ context.Context, profileID uuid.UUID,
) error {
	storage.retired = profileID
	return nil
}

func (storage *fakeStore) Claim(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Duration,
) (Claim, bool, error) {
	claim, found := storage.claim, storage.found
	storage.found = false
	return claim, found, nil
}

func (storage *fakeStore) Heartbeat(
	context.Context, Claim, time.Duration,
) (time.Time, error) {
	return time.Now().Add(LeaseTTL), nil
}

func (storage *fakeStore) Complete(
	_ context.Context, claim Claim, _ uuid.UUID, _ MigrationReceipt,
) error {
	storage.completed = claim
	return nil
}

func (storage *fakeStore) Fail(
	_ context.Context, _ Claim, failure Failure,
) error {
	storage.failure = failure
	return nil
}

func testExecutor(t *testing.T, store Store, migrator Migrator) *Executor {
	t.Helper()
	executor, err := New(Dependencies{Store: store, Migrator: migrator,
		Release: Release{ID: uuid.New(), DeploymentMode: "onpremise",
			RuntimeGeneration: 1, MigrationManifestSHA256: strings.Repeat("c", 64)},
		NewID: func() (uuid.UUID, error) { return uuid.New(), nil }})
	if err != nil {
		t.Fatal(err)
	}
	return executor
}

func validClaim() Claim {
	return Claim{JobID: uuid.New(), OperationID: uuid.New(), TenantID: uuid.New(),
		MigrationRunID: uuid.New(), LeaseID: uuid.New(), ProfileInstanceID: uuid.New(),
		ReleaseID: uuid.New(), EnvironmentID: uuid.New(), Origin: "new", Attempt: 1,
		SchemaName: "tenant_example", StepPlanSHA256: strings.Repeat("d", 64),
		DeploymentMode: "onpremise", RuntimeGeneration: 1, DatabaseGeneration: 1,
		MigrationManifestSHA256: strings.Repeat("c", 64)}
}
