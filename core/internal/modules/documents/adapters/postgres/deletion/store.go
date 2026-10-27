package deletion

import (
	"context"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const storageOwner = "documents"

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row
type rowNotFound func(error) bool
type scopeCheck func(context.Context) bool

type store struct {
	tenantID domain.ID
	queryRow queryRow
	notFound rowNotFound
	allowed  scopeCheck
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID.Valid() &&
		storage.queryRow != nil && storage.notFound != nil && storage.allowed != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}

var _ deletionapp.Store = (*store)(nil)
