package identity

import (
	"errors"
	"testing"
)

func TestOrganizationOwnerCommandsRequireCanonicalNonEmptySets(t *testing.T) {
	command := validOrganizationCreateCommand()
	if command.Validate() != nil {
		t.Fatal("valid organization owner set was rejected")
	}
	command.OwnerIDs = nil
	command.OwnerMembershipIDs = nil
	if !errors.Is(command.Validate(), ErrInvalidCommand) {
		t.Fatal("organization create accepted no owner")
	}
	command = validOrganizationCreateCommand()
	command.OwnerIDs = []IdentityID{{3}, {2}}
	if !errors.Is(command.Validate(), ErrInvalidCommand) {
		t.Fatal("organization create accepted a non-canonical owner set")
	}
	command = validOrganizationCreateCommand()
	command.OwnerMembershipIDs = []IdentityID{{4}, {4}}
	if !errors.Is(command.Validate(), ErrInvalidCommand) {
		t.Fatal("organization create accepted duplicate membership IDs")
	}
}

func TestOrganizationOwnerUpdateDistinguishesOmittedAndExactReplacement(t *testing.T) {
	create := validOrganizationCreateCommand()
	update := UpdateOrganizationCommand{
		ID: create.ID, ExpectedRevision: 1, Name: create.Name, Slug: create.Slug,
		ProfilePatch: []byte(`{}`), Evidence: create.Evidence,
	}
	if update.Validate() != nil {
		t.Fatal("organization update rejected an omitted owner set")
	}
	update.ReplaceOwners = true
	if !errors.Is(update.Validate(), ErrInvalidCommand) {
		t.Fatal("organization update accepted an empty exact owner set")
	}
	update.OwnerIDs = create.OwnerIDs
	update.OwnerMembershipIDs = create.OwnerMembershipIDs
	if update.Validate() != nil {
		t.Fatal("organization update rejected a canonical exact owner set")
	}
}

func validOrganizationCreateCommand() CreateOrganizationCommand {
	return CreateOrganizationCommand{
		ID: IdentityID{1}, Name: "Research", Slug: "research",
		OwnerIDs:           []IdentityID{{2}, {3}},
		OwnerMembershipIDs: []IdentityID{{4}, {5}},
		ProfileValues:      []byte(`{}`),
		Evidence: MutationEvidence{
			OperationID: IdentityID{6}, RequestSHA256: [32]byte{7},
			ActorID: IdentityID{8}, PermissionID: IdentityID{9},
			ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
		},
	}
}
