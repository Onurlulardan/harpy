package publication

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
	"time"
)

func TestActivateCapsCapabilityLifetimeAtThirtyDays(t *testing.T) {
	_, err := (&Application{}).Activate(context.Background(), platformtx.TenantScope{}, ActivateRequest{ReleaseID: publicationFixtureRelease, ExpectedRevision: 1, CapabilityTTL: 30*24*time.Hour + time.Microsecond, IdempotencyKey: "activate-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("activate error = %v", err)
	}
}
