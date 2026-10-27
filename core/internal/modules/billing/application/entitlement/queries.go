package entitlement

type Queries struct {
	store Store
	clock Clock
}

func NewQueries(store Store, clock Clock) (*Queries, error) {
	if store == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &Queries{store: store, clock: clock}, nil
}
