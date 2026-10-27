package licenseclock

import (
	"context"
	"fmt"

	licenseclockapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseclock"
)

type Row interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) Row
}

type Store interface{ licenseclockapp.Store }

type store struct{ database DatabaseAccess }

var _ Store = (*store)(nil)

func NewStore(database DatabaseAccess) (Store, error) {
	if database.QueryRow == nil {
		return nil, fmt.Errorf("billing license clock database is required")
	}
	return &store{database: database}, nil
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.database.QueryRow != nil &&
		ctx != nil && ctx.Err() == nil
}
