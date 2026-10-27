package providerapp

type CommandCommands struct {
	store  CommandStore
	nextID CommandIDSource
	clock  CommandClock
}

func NewCommandCommands(
	store CommandStore,
	nextID CommandIDSource,
	clock CommandClock,
) (*CommandCommands, error) {
	if store == nil || nextID == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &CommandCommands{store: store, nextID: nextID, clock: clock}, nil
}
