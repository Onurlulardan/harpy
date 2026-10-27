package policypostgres

import (
	"context"
	"testing"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAuthorizeBuildsProductionPolicyAndRejectsClosedScope(t *testing.T) {
	port, err := NewAuthorizePolicyPort()
	if err != nil {
		t.Fatal(err)
	}
	_, err = port.EvaluateQuery(
		context.Background(), transaction.TenantScope{}, restauthorization.PolicyRequest{},
	)
	if err == nil {
		t.Fatal("closed scope was accepted")
	}
}
