package authorization

import "testing"

func TestEvaluateCurrentActorUsesCurrentVersionWithoutTokenSnapshot(t *testing.T) {
	snapshot := testSnapshot(t, []BindingFact{testExactBinding(SubjectUser)}, nil)
	decision := evaluateCurrentActor(snapshot)
	permissionID, found := decision.PermissionID()
	if !decision.Allowed() || decision.AuthorizationVersion() != 9 ||
		!found || permissionID != testPermissionID {
		t.Fatalf("current actor decision = %+v", decision)
	}
}

func TestEvaluateCurrentActorFailsClosedAfterPermissionRevocation(t *testing.T) {
	binding := testExactBinding(SubjectUser)
	binding.BindingState = "revoked"
	snapshot := testSnapshot(t, []BindingFact{binding}, nil)
	decision := evaluateCurrentActor(snapshot)
	if decision.Allowed() || decision.Reason() != ReasonNoEffectivePermission {
		t.Fatalf("revoked current actor decision = %+v", decision)
	}
}
