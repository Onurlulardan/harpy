package identity

type IDQuery struct {
	ID       IdentityID
	Evidence QueryEvidence
}

func (query IDQuery) Validate() error {
	if query.ID == (IdentityID{}) || query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}

type ListQuery struct {
	AfterID  IdentityID
	Limit    int
	Evidence QueryEvidence
}

func (query ListQuery) Validate() error {
	if query.Limit < 1 || query.Limit > 100 || query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}
