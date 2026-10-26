package postrestore

import (
	"context"
	"errors"
	"testing"
)

func TestReconcileRejectsInvalidCommand(t *testing.T) {
	reconciler, err := NewReconciler(&postRestoreStoreFixture{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = reconciler.Reconcile(context.Background(), ReconcileCommand{})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("Reconcile() error = %v, want ErrDenied", err)
	}
}
