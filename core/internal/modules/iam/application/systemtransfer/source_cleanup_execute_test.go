package systemtransfer

import (
	"context"
	"errors"
	"testing"
)

func TestImportSourceCleanupPersistsExactAbsenceEvidence(t *testing.T) {
	target, deleted, now := importCleanupFixture()
	store := &importCleanupStoreStub{}
	objects := &importCleanupObjectStub{result: deleted}
	runner := &SourceCleanupRunner{store: store, objects: objects,
		clock: importCleanupValuesStub{now: now}, jitter: importCleanupJitterStub{}}
	if err := runner.execute(t.Context(), target); err != nil {
		t.Fatal(err)
	}
	if objects.key != target.ObjectKey || objects.version != target.ObjectVersion ||
		store.completion.Target != target || store.completion.Receipt != deleted {
		t.Fatalf("key=%q version=%q completion=%+v",
			objects.key, objects.version, store.completion)
	}
}

func TestImportSourceCleanupRetriesUnknownDeleteOutcome(t *testing.T) {
	target, _, now := importCleanupFixture()
	store := &importCleanupStoreStub{}
	objects := &importCleanupObjectStub{err: ErrSourceDeleteOutcomeUnknown}
	runner := &SourceCleanupRunner{store: store, objects: objects,
		clock: importCleanupValuesStub{now: now}, jitter: importCleanupJitterStub{value: 7}}
	if err := runner.execute(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	if store.retry.ErrorCode != SourceCleanupOutcomeUnknown ||
		store.retry.Target != target || store.retry.RetryAt.Sub(now) != 7 {
		t.Fatalf("retry=%+v", store.retry)
	}
	store.retryErr = ErrSourceCleanupStale
	if err := runner.execute(context.Background(), target); err != nil &&
		!errors.Is(err, ErrSourceCleanupStale) {
		t.Fatalf("stale retry error=%v", err)
	}
}

func TestImportSourceCleanupRefusesUnprovenAbsence(t *testing.T) {
	target, deleted, _ := importCleanupFixture()
	deleted.AbsenceProbeSHA256 = [32]byte{}
	if _, err := FinalizeSourceCleanup(
		t.Context(), &importCleanupStoreStub{}, target, deleted,
	); !errors.Is(err, ErrSourceDeleteOutcomeUnknown) {
		t.Fatalf("finalize error=%v", err)
	}
}
