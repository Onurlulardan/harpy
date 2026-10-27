package rbac

type Binding struct {
	ID           ID
	SubjectID    ID
	PermissionID ID
	State        string
	Revision     int64
}

type BindingCommand struct {
	ID               ID
	SubjectID        ID
	PermissionID     ID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

type BindingListQuery struct {
	SubjectID ID
	State     string
	Limit     int
	Evidence  QueryEvidence
}

func (command BindingCommand) ValidateGrant() error {
	if !command.ID.Valid() || !command.SubjectID.Valid() ||
		!command.PermissionID.Valid() || command.ExpectedRevision < 0 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (command BindingCommand) ValidateRevoke() error {
	if !command.ID.Valid() || !command.SubjectID.Valid() ||
		!command.PermissionID.Valid() || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (query BindingListQuery) Validate() error {
	if !query.SubjectID.Valid() || query.Limit < 1 || query.Limit > 200 ||
		(query.State != "active" && query.State != "revoked" && query.State != "all") ||
		query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}
