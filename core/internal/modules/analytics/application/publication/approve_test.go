package publication

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
	"time"
)

func TestApproveRequiresStepUpNoOlderThanTenMinutes(t *testing.T) {
	application := &Application{authorizer: &publicationFixtureAuthorizer{
		stepUp: publicationFixtureNow.Add(-10*time.Minute - time.Microsecond)},
		clock: publicationFixtureClock{publicationFixtureNow}}
	_, err := application.freshTransitionTime(context.Background(), platformtx.TenantScope{})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("step-up error = %v", err)
	}
}
