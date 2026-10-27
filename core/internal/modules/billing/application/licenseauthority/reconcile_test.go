package licenseauthority

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type reconcileStore struct {
	claim    RefreshClaim
	accepts  int
	failures int
	stored   VerifiedSnapshot
}

func (store *reconcileStore) AcquireRefreshLease(
	_ context.Context,
	_ RefreshCommand,
	_ []ComponentVersion,
) (RefreshClaim, error) {
	return store.claim, nil
}

func (store *reconcileStore) PersistVerifiedSnapshot(
	_ context.Context,
	_ RefreshClaim,
	snapshot VerifiedSnapshot,
) error {
	store.accepts++
	store.stored = snapshot
	return nil
}

func (store *reconcileStore) StoreRefreshFailure(
	context.Context,
	RefreshClaim,
	FailureKind,
) error {
	store.failures++
	return nil
}

type reconcileSource struct {
	document []byte
	err      error
}

func (source reconcileSource) IssueSnapshot(context.Context, IssueRequest) ([]byte, error) {
	return append([]byte(nil), source.document...), source.err
}

func TestReconcileAcceptsOnlyVerifiedHigherAuthority(t *testing.T) {
	fixture := newAuthorityFixture(t)
	command, store := reconcileCommand(fixture)
	reconciler := newReconcilerFixture(t, fixture, store, fixture.document(t))
	result, err := reconciler.Reconcile(context.Background(), command)
	if err != nil || result.Decision != SequenceAccept || store.accepts != 1 ||
		store.stored.Digest() == "" {
		t.Fatalf("verified snapshot was not accepted: result=%+v accepts=%d err=%v",
			result, store.accepts, err)
	}
	store.claim.Head = &AcceptedHead{
		LicenseID: fixture.binding.LicenseID, EnvironmentID: fixture.binding.EnvironmentID,
		Sequence: store.stored.Sequence(), Digest: store.stored.Digest(),
	}
	result, err = reconciler.Reconcile(context.Background(), command)
	if err != nil || result.Decision != SequenceNoOp || store.accepts != 2 {
		t.Fatalf("idempotent snapshot rewrote cache: result=%+v accepts=%d err=%v",
			result, store.accepts, err)
	}
}

func TestReconcilePreservesCacheOnRemoteAndVerificationFailure(t *testing.T) {
	fixture := newAuthorityFixture(t)
	command, store := reconcileCommand(fixture)
	reconciler := newReconcilerFixture(t, fixture, store, []byte(`{"bad":true}`))
	if _, err := reconciler.Reconcile(context.Background(), command); err == nil ||
		store.accepts != 0 || store.failures != 1 {
		t.Fatalf("invalid remote response mutated cache: accepts=%d err=%v", store.accepts, err)
	}
	reconciler.source = reconcileSource{err: errors.New("issuer unavailable")}
	if _, err := reconciler.Reconcile(context.Background(), command); err == nil ||
		store.accepts != 0 || store.failures != 2 {
		t.Fatalf("remote outage mutated cache: accepts=%d err=%v", store.accepts, err)
	}
}

func newReconcilerFixture(
	t *testing.T,
	fixture authorityFixture,
	store Store,
	document []byte,
) *Reconciler {
	t.Helper()
	reconciler, err := NewLicenseReconciler(
		store, reconcileSource{document: document}, fixture.resolver(),
		[]ComponentVersion{{Component: "core-worker", Version: "v2.0.0"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	return reconciler
}

func reconcileCommand(fixture authorityFixture) (RefreshCommand, *reconcileStore) {
	operationID := uuid.New()
	command := RefreshCommand{OperationID: operationID, Binding: fixture.binding}
	store := &reconcileStore{claim: RefreshClaim{
		OperationID: operationID, LeaseToken: uuid.New(), Binding: fixture.binding,
		Nonce: [32]byte{1}, DatabaseNow: fixture.now,
		LeaseExpiresAt: fixture.now.Add(time.Minute),
	}}
	return command, store
}
