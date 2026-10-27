package v1bridge

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	bridgeapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1bridge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type Database interface {
	QueryRow(context.Context, string, ...any) platformpostgres.Row
	Execute(context.Context, string, ...any) (platformpostgres.CommandResult, error)
}

type Enqueuer interface {
	EnqueueNotificationV1Bridge(context.Context, bridgeapp.Job) error
}

type Bound struct {
	Database Database
	Enqueuer Enqueuer
	TenantID uuid.UUID
}

type Source interface {
	BindNotificationV1Bridge(context.Context, api.Scope) (Bound, error)
}

type Binder struct{ source Source }

func NewBinder(source Source) (*Binder, error) {
	if source == nil {
		return nil, bridgeapp.ErrConfiguration
	}
	return &Binder{source: source}, nil
}

func (binder *Binder) BindBridge(ctx context.Context,
	scope api.Scope,
) (bridgeapp.Store, error) {
	identity, scoped := scope.(interface{ TenantID() uuid.UUID })
	if binder == nil || binder.source == nil || ctx == nil || scope == nil || !scoped {
		return nil, bridgeapp.ErrStorage
	}
	bound, err := binder.source.BindNotificationV1Bridge(ctx, scope)
	if err != nil || bound.Database == nil || bound.Enqueuer == nil ||
		bound.TenantID == uuid.Nil || bound.TenantID != identity.TenantID() {
		return nil, bridgeapp.ErrStorage
	}
	return &Store{database: bound.Database, enqueuer: bound.Enqueuer,
		tenantID: bound.TenantID}, nil
}

func classifyNotFound(err error) bool { return platformpostgres.RowNotFound(err) }

var _ bridgeapp.Binder = (*Binder)(nil)
