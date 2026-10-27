package authorization

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAuthorizerIsFixedRedactedAndRejectsClosedScope(t *testing.T) {
	port := &countingPolicyPort{}
	registry, err := NewPolicyRegistry([]PolicyRegistration{
		{Owner: "forms", Port: port},
		{Owner: "crm", Port: port},
	})
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := NewAuthorizer(registry)
	if err != nil || authorizer.Validate() != nil {
		t.Fatalf("NewAuthorizer() error = %v", err)
	}
	fixture := newAuthorizationQueryFixture(t)
	_, err = authorizer.Authorize(
		context.Background(), transaction.TenantScope{},
		authorizationRecordSpec(t, fixture),
	)
	if !errors.Is(err, errPolicyEvaluationContract) || port.calls != 0 {
		t.Fatalf("closed scope error=%v calls=%d", err, port.calls)
	}
	if fmt.Sprint(authorizer) != authorizationRedaction ||
		authorizer.LogValue().String() != authorizationRedaction {
		t.Fatal("authorizer formatting leaked")
	}
	if _, err := json.Marshal(authorizer); err == nil {
		t.Fatal("authorizer allowed direct JSON marshal")
	}
}

func TestNewAuthorizerRejectsInvalidRegistry(t *testing.T) {
	if _, err := NewAuthorizer(PolicyRegistry{}); err == nil {
		t.Fatal("invalid policy registry was accepted")
	}
}
