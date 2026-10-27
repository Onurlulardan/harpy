package subscriptionapp

import "context"

type CurrentStore interface {
	GetCurrent(context.Context) (CurrentSubscription, error)
}

type Queries struct {
	current CurrentStore
}

func NewQueries(current CurrentStore) (*Queries, error) {
	if current == nil {
		return nil, ErrQueryConfiguration
	}
	return &Queries{current: current}, nil
}
