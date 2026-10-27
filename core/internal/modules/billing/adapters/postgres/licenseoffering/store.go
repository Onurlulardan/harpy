package licenseoffering

import (
	"context"
	"fmt"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseoffering"
)

type Row interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) Row
}

type store struct{ database DatabaseAccess }

var _ offeringapp.Store = (*store)(nil)

func NewStore(database DatabaseAccess) (offeringapp.Store, error) {
	if database.QueryRow == nil {
		return nil, fmt.Errorf("billing license offering database is required")
	}
	return &store{database: database}, nil
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.database.QueryRow != nil &&
		ctx != nil && ctx.Err() == nil
}
