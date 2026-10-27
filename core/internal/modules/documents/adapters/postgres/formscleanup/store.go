package formscleanup

import (
	"context"

	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
)

const storageOwner = "documents"

type row interface{ Scan(...any) error }
type rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}
type queryRow func(context.Context, string, ...any) row
type query func(context.Context, string, ...any) (rows, error)
type rowNotFound func(error) bool
type scopeCheck func(context.Context) bool

type store struct {
	queryRow queryRow
	query    query
	notFound rowNotFound
	allowed  scopeCheck
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.queryRow != nil && storage.query != nil &&
		storage.notFound != nil && storage.allowed != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}

var _ formscleanupapp.Store = (*store)(nil)
