package providerapp

import "context"

type FinalizerStore interface {
	FinalizeCommand(context.Context, FinalizeCommandMutation) (Finalization, error)
}

type CommandFinalizer struct {
	store  FinalizerStore
	nextID CommandIDSource
	clock  CommandClock
}

func NewCommandFinalizer(
	store FinalizerStore,
	nextID CommandIDSource,
	clock CommandClock,
) (*CommandFinalizer, error) {
	if store == nil || nextID == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &CommandFinalizer{store: store, nextID: nextID, clock: clock}, nil
}
