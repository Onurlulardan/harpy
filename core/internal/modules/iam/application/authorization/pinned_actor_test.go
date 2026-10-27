package authorization

import "testing"

func testPinnedActorRequest(t *testing.T) PinnedActorRequest {
	t.Helper()
	request, err := NewPinnedActorRequest("documents.items", "import", 9)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestEvaluatePinnedActorAllowsCurrentActorAfterTokenRotation(t *testing.T) {
	binding := testExactBinding(SubjectUser)
	snapshot := testSnapshot(t, []BindingFact{binding}, nil)
	snapshot.actor.TokenVersion = 23

	decision := evaluatePinnedActor(testPinnedActorRequest(t), snapshot)
	permissionID, found := decision.PermissionID()
	if !decision.Allowed() || decision.TokenVersion() != 23 ||
		decision.AuthorizationVersion() != 9 || !found ||
		permissionID != binding.PermissionID {
		t.Fatalf("decision = %t/%s token=%d auth=%d permission=%s found=%t",
			decision.Allowed(), decision.Reason(), decision.TokenVersion(),
			decision.AuthorizationVersion(), permissionID, found)
	}
}

func TestEvaluatePinnedActorDeniesAuthorizationVersionDrift(t *testing.T) {
	snapshot := testSnapshot(t, []BindingFact{testExactBinding(SubjectUser)}, nil)
	snapshot.actor.AuthorizationVersion = 10

	decision := evaluatePinnedActor(testPinnedActorRequest(t), snapshot)
	if decision.Allowed() || decision.Reason() != ReasonAuthorizationStale {
		t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
	}
}

func TestEvaluatePinnedActorDeniesSuspendedActor(t *testing.T) {
	snapshot := testSnapshot(t, []BindingFact{testExactBinding(SubjectUser)}, nil)
	snapshot.actor.Status = "suspended"

	decision := evaluatePinnedActor(testPinnedActorRequest(t), snapshot)
	if decision.Allowed() || decision.Reason() != ReasonActorUnavailable {
		t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
	}
}

func TestEvaluatePinnedActorRequiresEffectiveGrant(t *testing.T) {
	decision := evaluatePinnedActor(
		testPinnedActorRequest(t), testSnapshot(t, nil, nil),
	)
	if decision.Allowed() || decision.Reason() != ReasonNoEffectivePermission {
		t.Fatalf("decision = %t/%s", decision.Allowed(), decision.Reason())
	}
}
