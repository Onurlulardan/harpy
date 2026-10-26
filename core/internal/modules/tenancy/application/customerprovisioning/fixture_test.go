package customerprovisioning

import (
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func provisionFixture(t *testing.T, mode tenantapi.DeploymentMode) (
	*Coordinator, *accessFixture, *storeFixture,
) {
	t.Helper()
	environment := "10000000-0000-4000-8000-000000000001"
	control := "20000000-0000-4000-8000-000000000001"
	maximum := int64(0)
	var limit *int64
	purpose := tenantapi.PrimaryTenantPurposeSaaS
	if mode == tenantapi.DeploymentModeOnPremise {
		limit = &maximum
		purpose = tenantapi.PrimaryTenantPurposeOnPremise
	}
	policy, err := tenantapi.NewDeploymentPolicy(environment, mode, limit, provisionNow)
	if err != nil {
		t.Fatal(err)
	}
	target, err := tenantapi.NewPrimaryTenantTarget(
		environment, mode, control, purpose, "control",
	)
	if err != nil {
		t.Fatal(err)
	}
	access, store, setup := &accessFixture{}, &storeFixture{}, &setupFixture{}
	coordinator, err := New(Dependencies{Policy: policyFixture{policy, target},
		Access: access, Offering: offeringFixture{}, Authority: authorityFixture{},
		Setup: setup, Store: store, Clock: fixedClock{}})
	if err != nil {
		t.Fatal(err)
	}
	return coordinator, access, store
}

func validCommand(caller CallerKind) Command {
	return Command{Caller: caller, IdempotencyKey: "customer-acme",
		Customer: CustomerRequest{Name: "Acme", Slug: "acme",
			FirstAdmin: FirstAdmin{DisplayName: "Ada Admin", Email: "ada@acme.test"}}}
}
