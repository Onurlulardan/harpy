package customerlifecycle

import (
	"context"
	"fmt"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

type Row interface{ Scan(...any) error }
type QueryRow func(context.Context, string, ...any) Row
type RowNotFound func(error) bool

type Store struct {
	queryRow QueryRow
	notFound RowNotFound
}

func NewStore(queryRow QueryRow, notFound RowNotFound) (*Store, error) {
	if queryRow == nil || notFound == nil {
		return nil, fmt.Errorf("customer lifecycle postgres query is required")
	}
	return &Store{queryRow: queryRow, notFound: notFound}, nil
}

func (store *Store) usable(ctx context.Context) bool {
	return store != nil && store.queryRow != nil && store.notFound != nil &&
		ctx != nil && ctx.Err() == nil
}

func (store *Store) classify(err error) error {
	if err == nil {
		return nil
	}
	if store != nil && store.notFound != nil && store.notFound(err) {
		return lifecycle.ErrNotFound
	}
	return lifecycle.ErrUnavailable
}

var (
	_ lifecycle.Store       = (*Store)(nil)
	_ lifecycle.WorkerStore = (*Store)(nil)
)
