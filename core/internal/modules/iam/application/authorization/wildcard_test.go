package authorization

import "testing"

func TestEvaluateAllowsOnlyProtectedPolicyWildcards(t *testing.T) {
	tests := []struct {
		kind   string
		reason Reason
	}{
		{"resource_all_actions", ReasonProtectedResourceActions},
		{"tenant_all", ReasonProtectedTenantAll},
	}
	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			fact := testWildcardBinding(test.kind)
			decision := evaluate(testRequest(t), testSnapshot(t, []BindingFact{fact}, nil))
			if !decision.Allowed() || decision.Reason() != test.reason {
				t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
			}
			fact.PolicyState = ""
			fact.PolicyRoleID = nil
			fact.PolicyPermissionID = nil
			fact.PolicyBindingID = nil
			decision = evaluate(testRequest(t), testSnapshot(t, []BindingFact{fact}, nil))
			if decision.Allowed() {
				t.Fatal("wildcard without active protected policy was allowed")
			}
		})
	}
}

func TestEvaluateRejectsSubstitutedTenantAllPolicyEvidence(t *testing.T) {
	fact := testWildcardBinding("tenant_all")
	other := *fact.PolicyPermissionID
	other[15]++
	fact.PolicyPermissionID = &other
	decision := evaluate(testRequest(t), testSnapshot(t, []BindingFact{fact}, nil))
	if decision.Allowed() || decision.Reason() != ReasonNoEffectivePermission {
		t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
	}
}
