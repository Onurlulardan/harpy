package authorization

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type countingPolicyPort struct {
	calls int
}

func (port *countingPolicyPort) EvaluateQuery(
	context.Context,
	transaction.TenantScope,
	PolicyRequest,
) (PolicyDecision, error) {
	port.calls++
	return PolicyDecision{}, nil
}

func TestEvaluatePolicyRequestsRejectsScopeOutsideCoordinatorCallback(t *testing.T) {
	port := &countingPolicyPort{}
	registry, err := NewPolicyRegistry([]PolicyRegistration{
		{Owner: "forms", Port: port},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = evaluatePolicyRequests(
		context.Background(), transaction.TenantScope{}, registry,
		[]PolicyRequest{mustPolicyRequest(t)},
	)
	if !errors.Is(err, errPolicyEvaluationContract) || port.calls != 0 {
		t.Fatalf("closed scope error=%v calls=%d", err, port.calls)
	}
}

func TestPolicyEvaluationFailureIsOpaqueAndPreservesCause(t *testing.T) {
	cause := errors.New("database unavailable with sensitive detail")
	failure := newPolicyEvaluationFailure(cause)
	if failure.Error() != errPolicyEvaluationContract.Error() {
		t.Fatalf("failure leaked detail: %q", failure.Error())
	}
	if !errors.Is(failure, errPolicyEvaluationContract) ||
		!errors.Is(failure, cause) {
		t.Fatal("failure did not preserve classifications")
	}
}
