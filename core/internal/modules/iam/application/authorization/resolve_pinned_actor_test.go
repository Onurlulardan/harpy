package authorization

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestResolvePinnedActorFailsClosedOutsideOwnedScope(t *testing.T) {
	authorizer, err := NewAuthorizer(unavailableEvidencePort{})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := authorizer.ResolvePinnedActor(
		context.Background(), transaction.TenantScope{}, testPinnedActorRequest(t),
	)
	if !errors.Is(err, ErrAuthorizationContract) || decision.Allowed() ||
		decision.Reason() != ReasonInvalidRequest {
		t.Fatalf("decision = %t/%s, err = %v",
			decision.Allowed(), decision.Reason(), err)
	}
}
