package authorization

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type unavailableEvidencePort struct{}

func (unavailableEvidencePort) Authorize(
	context.Context,
	transaction.TenantScope,
	Request,
) (Snapshot, error) {
	return Snapshot{}, ErrAuthorizationEvidence
}

func TestAuthorizeFailsClosedBeforeEvidenceOutsideOwnedScope(t *testing.T) {
	if _, err := NewAuthorizer(nil); err == nil {
		t.Fatal("nil evidence port was accepted")
	}
	authorizer, err := NewAuthorizer(unavailableEvidencePort{})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := authorizer.Authorize(
		context.Background(), transaction.TenantScope{}, testRequest(t),
	)
	if err == nil || decision.Allowed() || decision.Reason() != ReasonInvalidRequest {
		t.Fatalf("decision = %t/%s, err = %v", decision.Allowed(), decision.Reason(), err)
	}
}
