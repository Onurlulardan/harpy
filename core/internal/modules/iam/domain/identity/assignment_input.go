package identity

type GrantAssignmentCommand struct {
	ID             IdentityID
	UserID         IdentityID
	RoleID         IdentityID
	ScopeKind      RoleScopeKind
	ScopeID        IdentityID
	OrganizationID IdentityID
	MembershipID   IdentityID
	Evidence       MutationEvidence
}

func (command GrantAssignmentCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.UserID == (IdentityID{}) || command.RoleID == (IdentityID{}) ||
		!ValidRoleScope(command.ScopeKind, command.ScopeID, command.OrganizationID) ||
		(command.ScopeKind == RoleScopeTenant && command.MembershipID != (IdentityID{})) ||
		(command.ScopeKind == RoleScopeOrganization && command.MembershipID == (IdentityID{})) ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type UpdateAssignmentCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	State            AssignmentState
	Evidence         MutationEvidence
}

func (command UpdateAssignmentCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		(command.State != AssignmentActive && command.State != AssignmentSuspended) ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type RevokeAssignmentCommand struct {
	ID               IdentityID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

func (command RevokeAssignmentCommand) Validate() error {
	if command.ID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type GetAssignmentQuery IDQuery
type ListAssignmentsQuery ListQuery

func (query GetAssignmentQuery) Validate() error { return IDQuery(query).Validate() }
func (query ListAssignmentsQuery) Validate() error {
	return ListQuery(query).Validate()
}
