package campaign

type Commands struct {
	store  Store
	nextID IDGenerator
	clock  Clock
}

func NewCommands(store Store, nextID IDGenerator, clock Clock) (*Commands, error) {
	if store == nil || nextID == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &Commands{store: store, nextID: nextID, clock: clock}, nil
}
