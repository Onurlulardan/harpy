package postrestore

import (
	"context"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

func TestReconcileRejectsUnboundStore(t *testing.T) {
	_, err := (&Store{}).Reconcile(context.Background(), application.ReconcileCommand{})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("Reconcile() error = %v, want ErrUnavailable", err)
	}
}
