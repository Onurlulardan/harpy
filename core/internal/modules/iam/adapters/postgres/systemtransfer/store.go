package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

const storageOwner = "iam"

type row interface{ Scan(...any) error }
type rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type store struct {
	tenantID, actorID uuid.UUID
	command           bool
	queryRow          func(context.Context, string, ...any) row
	query             func(context.Context, string, ...any) (rows, error)
	allowed           func(context.Context) bool
	notFound          func(error) bool
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.actorID != uuid.Nil &&
		storage.queryRow != nil && storage.query != nil && storage.allowed != nil &&
		storage.notFound != nil && ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}

var _ app.ControlStore = (*store)(nil)
