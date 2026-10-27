package identity

import (
	"errors"
	"testing"
)

func TestReplaceUserRelationsCommandBindsCommandAndActorRules(t *testing.T) {
	userID := IdentityID{1}
	command := validReplaceUserRelationsCommand(t, userID)
	if command.Validate() != nil {
		t.Fatal("valid relation replacement was rejected")
	}
	command.ExpectedRevision = 2
	if !errors.Is(command.Validate(), ErrInvalidCommand) {
		t.Fatal("create relation replacement accepted a non-initial revision")
	}
	command.CommandKey = UserRelationUpdateCommandKey
	if command.Validate() != nil {
		t.Fatal("update relation replacement was rejected")
	}
	command.Evidence.ActorID = userID
	if !errors.Is(command.Validate(), ErrInvalidCommand) {
		t.Fatal("self relation replacement was accepted")
	}
}

func TestReplaceUserRelationsCommandAllowsProfileOnlySelfUpdate(t *testing.T) {
	userID := IdentityID{1}
	command := validReplaceUserRelationsCommand(t, userID)
	command.CommandKey = UserRelationUpdateCommandKey
	command.ExpectedRevision = 4
	command.Organizations = OrganizationSetPatch{}
	command.PatchProfile = true
	command.ProfilePatch = []byte("{}")
	command.Evidence.ActorID = userID
	if command.Validate() != nil {
		t.Fatal("profile-only self update was rejected")
	}
	command.PatchProfile = false
	command.ProfilePatch = nil
	if !errors.Is(command.Validate(), ErrInvalidCommand) {
		t.Fatal("empty graph mutation was accepted")
	}
}

func validReplaceUserRelationsCommand(
	t *testing.T,
	userID IdentityID,
) ReplaceUserRelationsCommand {
	t.Helper()
	organizations, err := NewOrganizationSetPatch(
		userID, true, []IdentityID{{3}},
	)
	if err != nil {
		t.Fatalf("organizations: %v", err)
	}
	return ReplaceUserRelationsCommand{
		UserID: userID, ExpectedRevision: 1,
		CommandKey:    UserRelationCreateCommandKey,
		Organizations: organizations,
		Evidence: MutationEvidence{
			OperationID: IdentityID{4}, RequestSHA256: [32]byte{5},
			ActorID: IdentityID{6}, PermissionID: IdentityID{7},
			ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
		},
	}
}
