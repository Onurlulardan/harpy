package resourceentitlement

import "time"

type Commands struct {
	store Store
	usage Usage
	clock Clock
}

func NewCommands(store Store, usage Usage, clock Clock) (*Commands, error) {
	if store == nil || usage == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &Commands{store: store, usage: usage, clock: clock}, nil
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1_000 == 0
}
