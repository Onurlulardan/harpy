package identity

type AddMembershipCommand struct {
	ID             IdentityID
	OrganizationID IdentityID
	UserID         IdentityID
	Kind           MembershipKind
	Evidence       MutationEvidence
}

func (command AddMembershipCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.OrganizationID == (IdentityID{}) ||
		command.UserID == (IdentityID{}) || !command.Kind.Valid() ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type UpdateMembershipCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	Kind             MembershipKind
	State            MembershipState
	Evidence         MutationEvidence
}

func (command UpdateMembershipCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 || !command.Kind.Valid() ||
		(command.State != MembershipActive && command.State != MembershipSuspended) ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type DeleteMembershipCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

func (command DeleteMembershipCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type GetMembershipQuery IDQuery
type ListMembershipsQuery ListQuery

func (query GetMembershipQuery) Validate() error { return IDQuery(query).Validate() }
func (query ListMembershipsQuery) Validate() error {
	return ListQuery(query).Validate()
}
