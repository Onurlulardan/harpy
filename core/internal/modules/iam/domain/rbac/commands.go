package rbac

type GetQuery struct {
	ID       ID
	Evidence QueryEvidence
}

type ListQuery struct {
	State    string
	Limit    int
	Evidence QueryEvidence
}

type UpdateCatalog struct {
	ID               ID
	DisplayName      string
	Description      string
	ExpectedRevision int64
	Evidence         MutationEvidence
}

type StateCommand struct {
	ID               ID
	ExpectedRevision int64
	ReceiptID        ID
	Evidence         MutationEvidence
}

func (query GetQuery) Validate() error {
	if !query.ID.Valid() || query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}

func (query ListQuery) Validate() error {
	if query.Limit < 1 || query.Limit > 200 ||
		(query.State != "active" && query.State != "retired" && query.State != "all") ||
		query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}

func (command StateCommand) Validate() error {
	if !command.ID.Valid() || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}
