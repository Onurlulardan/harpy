package recoveryanchor

import (
	"context"
	"testing"
)

type finalizerStore struct{ operationID string }

func (store *finalizerStore) FinalizeAuthority(
	_ context.Context,
	operationID string,
) error {
	store.operationID = operationID
	return nil
}

func TestFinalizeDelegatesExactOperation(t *testing.T) {
	store := &finalizerStore{}
	finalizer, err := NewAuthorityFinalizer(store)
	if err != nil {
		t.Fatal(err)
	}
	const operationID = "11111111-1111-4111-8111-111111111111"
	if err = finalizer.Finalize(context.Background(), operationID); err != nil {
		t.Fatal(err)
	}
	if store.operationID != operationID {
		t.Fatalf("finalized operation = %s", store.operationID)
	}
}
