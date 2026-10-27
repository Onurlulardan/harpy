package usage

import "context"

type CurrentUsageStore interface {
	Get(context.Context) (UsageSnapshot, error)
}

type Queries struct {
	current CurrentUsageStore
}

func NewQueries(current CurrentUsageStore) (*Queries, error) {
	if current == nil {
		return nil, ErrQueryConfiguration
	}
	return &Queries{current: current}, nil
}
