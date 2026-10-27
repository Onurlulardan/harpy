package invoice

type Finalizer struct {
	store  Store
	nextID IDSource
	clock  Clock
}

func NewFinalizer(store Store, nextID IDSource, clock Clock) (*Finalizer, error) {
	if store == nil || nextID == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &Finalizer{store: store, nextID: nextID, clock: clock}, nil
}
