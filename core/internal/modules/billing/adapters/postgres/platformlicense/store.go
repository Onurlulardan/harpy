package platformlicense

import (
	"context"
	"fmt"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
)

type Row interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) Row
}

type Store struct{ database DatabaseAccess }

var _ platformlicenseapp.Store = (*Store)(nil)

func NewStore(database DatabaseAccess) (*Store, error) {
	if database.QueryRow == nil {
		return nil, fmt.Errorf("billing platform license database is required")
	}
	return &Store{database: database}, nil
}

func (store *Store) usable(ctx context.Context) bool {
	return store != nil && store.database.QueryRow != nil &&
		ctx != nil && ctx.Err() == nil
}
