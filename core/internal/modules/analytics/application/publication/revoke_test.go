package publication

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestRevokeRequiresIdempotencyKey(t *testing.T) {
	_, err := (&Application{}).Revoke(context.Background(), platformtx.TenantScope{}, TransitionRequest{ReleaseID: publicationFixtureRelease, ExpectedRevision: 1})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("revoke error = %v", err)
	}
}
