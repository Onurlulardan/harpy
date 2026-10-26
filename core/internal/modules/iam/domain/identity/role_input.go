package identity

type CreateRoleCommand struct {
	ID             IdentityID
	ScopeKind      RoleScopeKind
	ScopeID        IdentityID
	OrganizationID IdentityID
	Name           string
	Description    string
	Default        bool
	ProfileValues  []byte
	Evidence       MutationEvidence
	profile        ProfileDocument
}

func (command CreateRoleCommand) Validate() error {
	if command.ID == (IdentityID{}) || !ValidRoleScope(
		command.ScopeKind, command.ScopeID, command.OrganizationID,
	) || !ValidDisplayName(command.Name) || len(command.Description) > 2000 ||
		!ValidProfileInput(command.ProfileValues) || command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (command CreateRoleCommand) WithProfile(profile ProfileDocument) CreateRoleCommand {
	command.profile = profile
	return command
}

func (command CreateRoleCommand) Profile() ProfileDocument { return command.profile }

type UpdateRoleCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	Name             string
	Description      string
	Default          bool
	ProfilePatch     []byte
	Evidence         MutationEvidence
	profile          ProfileDocument
}

func (command UpdateRoleCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		!ValidDisplayName(command.Name) || len(command.Description) > 2000 ||
		!ValidProfileInput(command.ProfilePatch) || command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (command UpdateRoleCommand) WithProfile(profile ProfileDocument) UpdateRoleCommand {
	command.profile = profile
	return command
}

func (command UpdateRoleCommand) Profile() ProfileDocument { return command.profile }

type RetireRoleCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

func (command RetireRoleCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type GetRoleQuery IDQuery
type ListRolesQuery ListQuery

func (query GetRoleQuery) Validate() error   { return IDQuery(query).Validate() }
func (query ListRolesQuery) Validate() error { return ListQuery(query).Validate() }
