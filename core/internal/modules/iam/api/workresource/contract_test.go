package workresource

import "testing"

func TestRequestAndDecisionContracts(t *testing.T) {
	if !(Request{Resource: "forms.definition", Action: "read"}).Valid() {
		t.Fatal("expected canonical Forms definition read permission")
	}
	if (Request{Resource: "Forms Definition", Action: "read"}).Valid() {
		t.Fatal("expected display-name resource to be rejected")
	}
	if !(Decision{Allowed: false, AuthorizationRevision: 9,
		AuthorizationClock: 11}).Valid() {
		t.Fatal("expected a denied current decision with evidence to be valid")
	}
	if (Decision{Allowed: true}).Valid() {
		t.Fatal("expected evidence-free allow to be rejected")
	}
}
