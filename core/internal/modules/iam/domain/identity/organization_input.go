package identity

type CreateOrganizationCommand struct {
	ID                 IdentityID
	Name               string
	Slug               string
	ParentID           IdentityID
	OwnerIDs           []IdentityID
	OwnerMembershipIDs []IdentityID
	ProfileValues      []byte
	Evidence           MutationEvidence
	profile            ProfileDocument
}

func (command CreateOrganizationCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ParentID == command.ID ||
		!ValidDisplayName(command.Name) || !ValidSlug(command.Slug) ||
		!validOrganizationOwnerSet(command.OwnerIDs, command.OwnerMembershipIDs, true) ||
		!ValidProfileInput(command.ProfileValues) || command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (command CreateOrganizationCommand) WithProfile(profile ProfileDocument) CreateOrganizationCommand {
	command.profile = profile
	return command
}

func (command CreateOrganizationCommand) Profile() ProfileDocument { return command.profile }

type UpdateOrganizationCommand struct {
	ID                 IdentityID
	ExpectedRevision   int64
	Name               string
	Slug               string
	ParentID           IdentityID
	ReplaceOwners      bool
	OwnerIDs           []IdentityID
	OwnerMembershipIDs []IdentityID
	ProfilePatch       []byte
	Evidence           MutationEvidence
	profile            ProfileDocument
}

func (command UpdateOrganizationCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		command.ParentID == command.ID || !ValidDisplayName(command.Name) ||
		!ValidSlug(command.Slug) || !ValidProfileInput(command.ProfilePatch) ||
		!validOrganizationOwnerReplacement(command) ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (command UpdateOrganizationCommand) WithProfile(profile ProfileDocument) UpdateOrganizationCommand {
	command.profile = profile
	return command
}

func (command UpdateOrganizationCommand) Profile() ProfileDocument { return command.profile }

type OrganizationStateCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

func (command OrganizationStateCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type GetOrganizationQuery IDQuery
type ListOrganizationsQuery ListQuery

func (query GetOrganizationQuery) Validate() error { return IDQuery(query).Validate() }
func (query ListOrganizationsQuery) Validate() error {
	return ListQuery(query).Validate()
}
