package rbac

type ProtectedAdminPolicy struct {
	RoleID                   ID
	TenantAllPermissionID    ID
	RolePermissionBindingID  ID
	MinimumActiveAssignments int
	AuthorizationEpoch       int64
	Revision                 int64
}

type UpdatePolicyCommand struct {
	PolicyID                 ID
	MinimumActiveAssignments int
	ExpectedRevision         int64
	Evidence                 MutationEvidence
}

func (command UpdatePolicyCommand) Validate() error {
	if !command.PolicyID.Valid() || command.MinimumActiveAssignments < 1 ||
		command.ExpectedRevision < 1 || command.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}
