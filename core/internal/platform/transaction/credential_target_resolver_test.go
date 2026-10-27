package transaction

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func TestCredentialResolverBinderRejectsMissingFactory(t *testing.T) {
	if _, err := NewCredentialResolverBinder(nil); err == nil {
		t.Fatal("nil resolver factory was accepted")
	}
	binder, err := NewCredentialResolverBinder(
		func(postgres.BoundExecutor) CredentialTargetResolver { return nil },
	)
	if err != nil || !binder.valid || binder.binder.Owner() != credentialBindingOwner {
		t.Fatalf("sealed resolver binder = %+v, %v", binder, err)
	}
}

func TestCredentialScopeRequestRequiresLookupAndSealedResolver(t *testing.T) {
	lookup, err := NewLoginCredentialLookup(
		"iam.login", testCredentialTenantID, CredentialEmailIdentifier,
		testCredentialKeyID, [32]byte{1},
	)
	if err != nil {
		t.Fatal(err)
	}
	binder, err := NewCredentialResolverBinder(
		func(postgres.BoundExecutor) CredentialTargetResolver { return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if !(CredentialScopeRequest{Lookup: lookup, Resolver: binder}).valid() {
		t.Fatal("valid credential request was rejected")
	}
	if (CredentialScopeRequest{Lookup: lookup}).valid() ||
		(CredentialScopeRequest{Resolver: binder}).valid() {
		t.Fatal("partial credential request was accepted")
	}
}
