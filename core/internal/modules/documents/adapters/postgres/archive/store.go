package archive

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
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
type execute func(context.Context, string, ...any) (int64, error)
type rowNotFound func(error) bool
type scopeCheck func(context.Context) bool

type store struct {
	actorID, tenantID domain.ID
	command, job      bool
	queryRow          queryRow
	query             queryMany
	execute           execute
	notFound          rowNotFound
	allowed           scopeCheck
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID.Valid() &&
		(storage.job || storage.actorID.Valid()) && storage.queryRow != nil &&
		storage.query != nil && storage.execute != nil && storage.notFound != nil &&
		storage.allowed != nil && ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}

var _ archiveapp.Store = (*store)(nil)
