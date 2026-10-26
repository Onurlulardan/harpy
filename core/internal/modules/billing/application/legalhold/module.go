package legalhold

import (
	"context"
	"reflect"
)

type Commands struct {
	store  Store
	nextID IDSource
	clock  Clock
}

func NewCommands(store Store, nextID IDSource, clock Clock) (*Commands, error) {
	if nilStore(store) || nextID == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &Commands{store: store, nextID: nextID, clock: clock}, nil
}

func (commands *Commands) Place(
	ctx context.Context,
	request PlaceRequest,
) (Snapshot, error) {
	return commands.Create(ctx, request)
}

func nilStore(store Store) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
