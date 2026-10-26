package discount

import "context"

type EligibilityStore interface {
	FindEligible(context.Context, EligibilityLookup) (Eligibility, error)
}

type Queries struct {
	store EligibilityStore
	clock Clock
}

func NewQueries(store EligibilityStore, clock Clock) (*Queries, error) {
	if store == nil || clock == nil {
		return nil, ErrQueryConfiguration
	}
	return &Queries{store: store, clock: clock}, nil
}
