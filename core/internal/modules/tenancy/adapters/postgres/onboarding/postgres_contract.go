package onboarding

import (
	"context"
	"errors"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

var errUnavailable = errors.New("tenancy onboarding postgres unavailable")

type Row interface{ Scan(...any) error }

type QueryRow func(context.Context, string, ...any) Row
type Execute func(context.Context, string, ...any) error

type postgresStore struct {
	queryRow QueryRow
	execute  Execute
}

func NewStore(queryRow QueryRow, execute Execute) (onboardingapp.OperationStore, error) {
	if queryRow == nil || execute == nil {
		return nil, errUnavailable
	}
	return &postgresStore{queryRow: queryRow, execute: execute}, nil
}

func (storage *postgresStore) usable(ctx context.Context) bool {
	return storage != nil && storage.queryRow != nil && storage.execute != nil &&
		ctx != nil && ctx.Err() == nil
}

var _ onboardingapp.OperationStore = (*postgresStore)(nil)
