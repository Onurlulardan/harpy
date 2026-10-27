package download

import (
	"context"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const storageOwner = "documents"

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row
type rowNotFound func(error) bool
type scopeCheck func(context.Context) bool

type store struct {
	actorID, tenantID domain.ID
	queryRow          queryRow
	notFound          rowNotFound
	allowed           scopeCheck
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.actorID.Valid() && storage.tenantID.Valid() &&
		storage.queryRow != nil && storage.notFound != nil && storage.allowed != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}

var _ downloadapp.Store = (*store)(nil)
