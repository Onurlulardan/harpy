package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

const exportWorkerOperation = "iam.record_export.execute"

type OutboundStore struct {
	bindOwned func(tenantScope) (boundDatabase, error)
}

type exportWorkerBound struct {
	database          boundDatabase
	tenantID, actorID uuid.UUID
	allowed           func(context.Context) bool
}

func (store *OutboundStore) BindExportWorker(ctx context.Context,
	scope tenantScope,
) (app.ExportWorkerStore, error) {
	if store == nil || store.bindOwned == nil ||
		!exportWorkerScopeUsable(ctx, scope) {
		return nil, app.ErrExportUnavailable
	}
	bound, err := store.bindOwned(scope)
	if err != nil {
		return nil, app.ErrExportUnavailable
	}
	result := &exportWorkerBound{database: bound,
		tenantID: scope.TenantID(), actorID: scope.ActorID()}
	result.allowed = func(next context.Context) bool {
		return exportWorkerScopeUsable(next, scope)
	}
	return result, nil
}

func (store *exportWorkerBound) usable(ctx context.Context,
	attempt app.ExportAttempt,
) bool {
	return store != nil && store.allowed != nil && store.allowed(ctx) &&
		store.tenantID == attempt.TenantID && store.actorID == attempt.ActorID &&
		attempt.Validate() == nil
}

func (store *exportWorkerBound) schema(ctx context.Context) (string, error) {
	schema, err := store.database.CurrentSchema(ctx)
	if err != nil {
		return "", app.ErrExportUnavailable
	}
	return schema, nil
}

var _ app.ExportWorkerStoreBinder = (*OutboundStore)(nil)
var _ app.ExportWorkerStore = (*exportWorkerBound)(nil)
