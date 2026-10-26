package actorbridgeissue

type Resolver struct {
	store Store
	valid bool
}

func NewResolver(store Store) (*Resolver, error) {
	if store == nil {
		return nil, ErrUnavailable
	}
	return &Resolver{store: store, valid: true}, nil
}
