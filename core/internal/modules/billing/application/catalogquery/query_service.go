package catalogquery

type Queries struct{ store Store }

func New(store Store) (*Queries, error) {
	if store == nil {
		return nil, ErrConfiguration
	}
	return &Queries{store: store}, nil
}
