package customerprovisioning

import (
	"context"
	"fmt"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
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
		return nil, fmt.Errorf("customer provisioning postgres query is required")
	}
	return &Store{queryRow: queryRow, notFound: notFound}, nil
}

func (store *Store) usable(ctx context.Context) bool {
	return store != nil && store.queryRow != nil && store.notFound != nil &&
		ctx != nil && ctx.Err() == nil
}

var _ provisioning.OperationStore = (*Store)(nil)
