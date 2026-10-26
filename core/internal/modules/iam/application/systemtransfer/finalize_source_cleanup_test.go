package systemtransfer

import (
	"context"
	"errors"
	"testing"
)

func TestFinalizeSourceCleanupPersistsExactAbsenceEvidence(t *testing.T) {
	target, deleted, _ := importCleanupFixture()
	store := &importCleanupStoreStub{}
	if _, err := FinalizeSourceCleanup(
		context.Background(), store, target, deleted,
	); err != nil {
		t.Fatal(err)
	}
	if store.completion.Target != target || store.completion.Receipt != deleted {
		t.Fatalf("completion = %#v", store.completion)
	}
}

func TestFinalizeSourceCleanupRejectsMissingAbsenceProof(t *testing.T) {
	target, deleted, _ := importCleanupFixture()
	deleted.AbsenceProbeSHA256 = [32]byte{}
	_, err := FinalizeSourceCleanup(
		context.Background(), &importCleanupStoreStub{}, target, deleted,
	)
	if !errors.Is(err, ErrSourceDeleteOutcomeUnknown) {
		t.Fatalf("error = %v", err)
	}
}
