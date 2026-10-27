package effectexecution_test

import (
	"context"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func TestUnknownResultNeverCarriesAReceipt(t *testing.T) {
	result := effectexecution.Unknown()
	if !result.IsUnknown() {
		t.Fatal("unknown certainty was lost")
	}
	if _, ok := result.Receipt(); ok {
		t.Fatal("unknown provider outcome exposed a definitive receipt")
	}
}

func TestProviderCannotMutateCallerRequestIdentity(t *testing.T) {
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	provider := providerStub{identity: identity, handler: func(request effectsv1.Request) effectexecution.Result {
		*request.TenantID = "99999999-9999-4999-8999-999999999999"
		return effectexecution.Unknown()
	}}
	registry, _ := effectexecution.NewRegistry([]effectexecution.Provider{provider})
	resolved, _ := registry.Resolve(identity)
	request := requestFixture(t)
	if _, err := resolved.Invoke(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if request.TenantID == nil || *request.TenantID != tenantID {
		t.Fatal("provider mutated caller request tenant identity")
	}
}

func TestDuplicateProviderOperationIsRejected(t *testing.T) {
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	first := providerStub{identity: identity}
	if _, err := effectexecution.NewRegistry([]effectexecution.Provider{first, first}); err == nil {
		t.Fatal("duplicate provider operation was accepted")
	}
}
