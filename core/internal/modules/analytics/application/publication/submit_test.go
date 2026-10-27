package publication

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestSubmitRequiresIdempotencyKey(t *testing.T) {
	_, err := (&Application{}).Submit(context.Background(), platformtx.TenantScope{}, TransitionRequest{ReleaseID: publicationFixtureRelease, ExpectedRevision: 1})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("submit error = %v", err)
	}
}
