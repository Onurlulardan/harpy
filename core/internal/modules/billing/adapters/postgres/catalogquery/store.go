package catalogquery

import (
	"context"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}

type queryRows func(context.Context, string, ...any) (rows, error)
type scopeCheck func(context.Context) bool

type store struct {
	query   queryRows
	allowed scopeCheck
}

var _ catalogapp.Store = (*store)(nil)

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.query != nil && storage.allowed != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}
