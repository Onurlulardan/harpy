package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

const storageOwner = "analytics"

type row interface{ Scan(...any) error }
type rowNotFound func(error) bool

type store struct {
	tenantID, actorID uuid.UUID
	command           bool
	queryRow          func(context.Context, string, ...any) row
	allowed           func(context.Context) bool
	notFound          rowNotFound
}

func (storage *store) usable(ctx context.Context, command bool) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.actorID != uuid.Nil &&
		storage.queryRow != nil && storage.allowed != nil && storage.notFound != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx) && (!command || storage.command)
}

var _ exportapp.Store = (*store)(nil)
