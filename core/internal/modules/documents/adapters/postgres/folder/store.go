package folder

import (
	"context"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const storageOwner = "documents"

type row interface{ Scan(...any) error }
type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}

type queryRow func(context.Context, string, ...any) row
type queryMany func(context.Context, string, ...any) (rows, error)
type rowNotFound func(error) bool
type scopeCheck func(context.Context) bool

type store struct {
	actorID, tenantID domain.ID
	command           bool
	queryRow          queryRow
	query             queryMany
	notFound          rowNotFound
	allowed           scopeCheck
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.actorID.Valid() && storage.tenantID.Valid() &&
		storage.queryRow != nil && storage.query != nil && storage.notFound != nil &&
		storage.allowed != nil && ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}

var _ folderapp.Store = (*store)(nil)
