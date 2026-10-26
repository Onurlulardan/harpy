package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

const storageOwner = "analytics"

type row interface{ Scan(...any) error }

type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}

type queryRow func(context.Context, string, ...any) row
type queryMany func(context.Context, string, ...any) (rows, error)
type scopeCheck func(context.Context) bool
type rowNotFound func(error) bool

type store struct {
	tenantID  uuid.UUID
	actorID   uuid.UUID
	command   bool
	operation string
	queryRow  queryRow
	query     queryMany
	allowed   scopeCheck
	notFound  rowNotFound
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil &&
		storage.actorID != uuid.Nil && storage.queryRow != nil &&
		storage.query != nil && storage.allowed != nil && storage.notFound != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}

var _ queryapp.Store = (*store)(nil)
