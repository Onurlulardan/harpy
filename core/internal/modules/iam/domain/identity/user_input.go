package identity

type CreateUserCommand struct {
	ID                  IdentityID
	IdentifierID        IdentityID
	AccessModeID        IdentityID
	DefaultAssignmentID IdentityID
	Status              UserStatus
	ProfileValues       []byte
	Directory           DirectoryProjection
	Evidence            MutationEvidence
	profile             ProfileDocument
}

func (command CreateUserCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.IdentifierID == (IdentityID{}) ||
		command.AccessModeID == (IdentityID{}) || command.DefaultAssignmentID == (IdentityID{}) ||
		command.Status != UserActive || !command.Directory.Valid() ||
		!ValidProfileInput(command.ProfileValues) || command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (command CreateUserCommand) WithProfile(profile ProfileDocument) CreateUserCommand {
	command.profile = profile
	return command
}

func (command CreateUserCommand) Profile() ProfileDocument { return command.profile }

type UpdateUserCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	ProfilePatch     []byte
	Evidence         MutationEvidence
	profile          ProfileDocument
}

func (command UpdateUserCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		!ValidProfileInput(command.ProfilePatch) || command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

func (command UpdateUserCommand) WithProfile(profile ProfileDocument) UpdateUserCommand {
	command.profile = profile
	return command
}

func (command UpdateUserCommand) Profile() ProfileDocument { return command.profile }

type UserStateCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

func (command UserStateCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type GetUserQuery IDQuery
type ListUsersQuery ListQuery

func (query GetUserQuery) Validate() error   { return IDQuery(query).Validate() }
func (query ListUsersQuery) Validate() error { return ListQuery(query).Validate() }
