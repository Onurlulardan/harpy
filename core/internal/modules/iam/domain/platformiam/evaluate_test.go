package platformiam

import "testing"

func TestEvaluateQueryAndDecisionFailClosed(t *testing.T) {
	query := EvaluateQuery{
		ActorID: fixtureID(2), Capability: "platform.instance.read",
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 2,
		ExpectedPlatformClock: 3,
	}
	if err := query.Validate(); err != nil {
		t.Fatal(err)
	}
	decision := Decision{
		Allowed: true, Capability: query.Capability, DeploymentMode: "onpremise",
		SubjectID: fixtureID(3), PlatformClock: 3, AuthorizationVersion: 2,
		Reason: "granted",
	}
	if err := decision.Validate(query); err != nil {
		t.Fatal(err)
	}
	decision.DeploymentMode = ""
	if err := decision.Validate(query); err == nil {
		t.Fatal("missing mode accepted")
	}
	decision.DeploymentMode = "saas"
	decision.Reason = "not_granted"
	if err := decision.Validate(query); err == nil {
		t.Fatal("inconsistent decision accepted")
	}
	decision.Allowed = false
	decision.SubjectID = ID{}
	if err := decision.Validate(query); err != nil {
		t.Fatalf("denied non-platform user was rejected: %v", err)
	}
}
