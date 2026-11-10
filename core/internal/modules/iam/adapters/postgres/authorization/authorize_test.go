package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAuthorizePostgresRejectsUnownedScope(t *testing.T) {
	authorizer, err := NewAuthorizer()
	if err != nil {
		t.Fatal(err)
	}
	request, err := authorization.NewRequest("iam.users", "read", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authorizer.Authorize(
		context.Background(), platformtx.TenantScope{}, request,
	); err == nil {
		t.Fatal("authorization read accepted an unowned tenant scope")
	}
}

func TestAuthorizerConstructorsSealExactTransactionOwner(t *testing.T) {
	iamPort, err := NewAuthorizer()
	if err != nil {
		t.Fatal(err)
	}
	iam, iamOK := iamPort.(*authorizer)
	if !iamOK || iam.binder.Owner() != "iam" {
		t.Fatal("authorization ports do not seal their exact transaction owners")
	}
}
