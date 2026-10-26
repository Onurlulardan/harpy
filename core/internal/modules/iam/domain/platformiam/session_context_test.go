package platformiam

import "testing"

func TestSessionContextSeparatesPrimaryAndCustomerAuthority(t *testing.T) {
	query := SessionContextQuery{
		ActorID: fixtureID(1), ExpectedTokenVersion: 2,
		ExpectedAuthorizationVersion: 3,
	}
	if err := query.Validate(); err != nil {
		t.Fatal(err)
	}
	primary := SessionContext{
		DeploymentMode: "onpremise", TenantPurpose: "onpremise_primary",
		PlatformClock: 4, AuthorizationVersion: 3,
	}
	if err := primary.Validate(); err != nil {
		t.Fatal(err)
	}
	customer := SessionContext{
		DeploymentMode: "saas", TenantPurpose: "saas_customer",
		PlatformClock: 0, AuthorizationVersion: 3,
	}
	if err := customer.Validate(); err != nil {
		t.Fatal(err)
	}
	customer.SubjectID = fixtureID(2)
	if err := customer.Validate(); err == nil {
		t.Fatal("customer context accepted a platform subject")
	}
}
