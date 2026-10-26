package adoption

import (
	"context"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func TestReconcileFailsClosedWithoutAdopterStore(t *testing.T) {
	_, err := (*Store)(nil).AdoptionReconcile(context.Background(), application.LeaseCommand{})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("reconcile error = %v, want unavailable", err)
	}
}
