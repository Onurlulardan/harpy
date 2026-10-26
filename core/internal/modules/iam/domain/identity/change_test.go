package identity

import "testing"

func TestIdentityChangeRequiresMatchingResourceAndResult(t *testing.T) {
	evidence := MutationEvidence{
		OperationID: IdentityID{1}, ActorID: IdentityID{2},
		PermissionID: IdentityID{3}, RequestSHA256: [32]byte{4},
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	change := NewChange(evidence, IdentityID{5}, "user", "iam.identity.user.update")
	if change.Validate() != nil {
		t.Fatal("valid identity change was rejected")
	}
	change.ResourceKind = "role"
	if change.Validate() == nil {
		t.Fatal("cross-resource identity command was accepted")
	}
}
