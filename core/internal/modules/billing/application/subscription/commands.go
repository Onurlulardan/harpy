package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type IDGenerator func() (uuid.UUID, error)
type Clock func() time.Time

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

func (commands *Commands) identifier() (uuid.UUID, error) {
	if commands == nil || commands.nextID == nil {
		return uuid.Nil, ErrConfiguration
	}
	identifier, err := commands.nextID()
	if err != nil || !validV7(identifier) {
		return uuid.Nil, ErrConfiguration
	}
	return identifier, nil
}

func (commands *Commands) instant() (time.Time, error) {
	if commands == nil || commands.clock == nil {
		return time.Time{}, ErrConfiguration
	}
	value := commands.clock()
	if !validInstant(value) {
		return time.Time{}, ErrConfiguration
	}
	return value, nil
}
