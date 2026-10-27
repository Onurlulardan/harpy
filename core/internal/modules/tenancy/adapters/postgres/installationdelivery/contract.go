package installationdelivery

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

type Row interface{ Scan(...any) error }
type QueryRow func(context.Context, string, ...any) Row

type store struct{ queryRow QueryRow }

func NewStore(queryRow QueryRow) (*store, error) {
	if queryRow == nil {
		return nil, installationdelivery.ErrUnavailable
	}
	return &store{queryRow: queryRow}, nil
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.queryRow != nil && ctx != nil && ctx.Err() == nil
}
