package authorization

import "testing"

func TestEvaluateDeniesUnavailableOrStaleContext(t *testing.T) {
	request := testRequest(t)
	tests := []struct {
		name   string
		mutate func(*Snapshot)
		reason Reason
	}{
		{"missing_actor", func(value *Snapshot) { value.actor = nil }, ReasonActorUnavailable},
		{"inactive_actor", func(value *Snapshot) { value.actor.Status = "suspended" }, ReasonActorUnavailable},
		{"stale_token", func(value *Snapshot) { value.actor.TokenVersion++ }, ReasonTokenStale},
		{"stale_authorization", func(value *Snapshot) { value.actor.AuthorizationVersion++ }, ReasonAuthorizationStale},
		{"missing_resource", func(value *Snapshot) { value.resource = nil }, ReasonResourceUnavailable},
		{"retired_resource", func(value *Snapshot) { value.resource.State = "retired" }, ReasonResourceUnavailable},
		{"missing_action", func(value *Snapshot) { value.action = nil }, ReasonActionUnavailable},
		{"retired_action", func(value *Snapshot) { value.action.State = "retired" }, ReasonActionUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := testSnapshot(t, nil, nil)
			test.mutate(&snapshot)
			decision := evaluate(request, snapshot)
			if decision.Allowed() || decision.Reason() != test.reason {
				t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
			}
		})
	}
}

func TestEvaluateUnionsExactSubjectPathsDeterministically(t *testing.T) {
	tests := []struct {
		path   SubjectPath
		reason Reason
	}{
		{SubjectUser, ReasonDirectUserExact},
		{SubjectRole, ReasonRoleExact},
		{SubjectOrganization, ReasonOrganizationExact},
	}
	for _, test := range tests {
		t.Run(string(test.path), func(t *testing.T) {
			binding := testExactBinding(test.path)
			snapshot := testSnapshot(t, []BindingFact{binding}, nil)
			decision := evaluate(testRequest(t), snapshot)
			if !decision.Allowed() || decision.Reason() != test.reason ||
				decision.AuthorizationClock() != 11 || decision.AuthorizationVersion() != 9 {
				t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
			}
			permissionID, found := decision.PermissionID()
			if !found || permissionID != binding.PermissionID {
				t.Fatalf("permission ID = %s, %t", permissionID, found)
			}
		})
	}
	all := []BindingFact{
		testExactBinding(SubjectOrganization), testExactBinding(SubjectRole),
		testExactBinding(SubjectUser),
	}
	decision := evaluate(testRequest(t), testSnapshot(t, all, nil))
	permissionID, found := decision.PermissionID()
	if decision.Reason() != ReasonDirectUserExact || !found ||
		permissionID != all[2].PermissionID {
		t.Fatalf("deterministic reason = %s", decision.Reason())
	}
}

func TestEvaluateRejectsInactiveRelationsAndParentInheritance(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*BindingFact)
	}{
		{"revoked_binding", func(value *BindingFact) { value.BindingState = "revoked" }},
		{"suspended_assignment", func(value *BindingFact) { value.AssignmentState = "suspended" }},
		{"retired_role", func(value *BindingFact) { value.RoleStatus = "retired" }},
		{"suspended_membership", func(value *BindingFact) { value.MembershipState = "suspended" }},
		{"suspended_organization", func(value *BindingFact) { value.OrganizationStatus = "suspended" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fact := testOrganizationRoleBinding()
			test.mutate(&fact)
			decision := evaluate(testRequest(t), testSnapshot(t, []BindingFact{fact}, nil))
			if decision.Allowed() || decision.Reason() != ReasonNoEffectivePermission {
				t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
			}
		})
	}
	decision := evaluate(testRequest(t), testSnapshot(t, nil, nil))
	if decision.Allowed() {
		t.Fatal("organization parent without a direct membership inherited access")
	}
}
