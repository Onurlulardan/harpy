package tenantmigration

import (
	"context"

	tenantapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantmigration"
)

type Row interface{ Scan(...any) error }

type QueryRow func(context.Context, string, ...any) Row

type repository struct{ queryRow QueryRow }

func NewRepository(queryRow QueryRow) (tenantapp.Repository, error) {
	if queryRow == nil {
		return nil, tenantapp.ErrUnavailable
	}
	return &repository{queryRow: queryRow}, nil
}

func (storage *repository) usable(ctx context.Context) bool {
	return storage != nil && storage.queryRow != nil && ctx != nil && ctx.Err() == nil
}

var _ tenantapp.Repository = (*repository)(nil)
