package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Binder struct{}

//go:embed sql/tenant/get_operation_actor.sql
var getOperationActorSQL string

type JobSnapshot struct {
	Operation catalogdomain.Operation
	ActorID   uuid.UUID
}

type JobBinding struct {
	binder platformtx.JobOwnerBinder[*jobReader]
}

type jobReader struct {
	queryRow func(context.Context, string, ...any) row
}

func NewBinder() *Binder { return &Binder{} }

func NewJobBinding() (JobBinding, error) {
	binder, err := platformtx.NewJobOwnerBinder(
		storageOwner,
		func(database platformtx.JobBoundExecutor) *jobReader {
			return &jobReader{queryRow: func(
				ctx context.Context, statement string, arguments ...any,
			) row {
				return database.QueryRow(ctx, statement, arguments...)
			}}
		},
	)
	if err != nil {
		return JobBinding{}, catalogapp.ErrUnavailable
	}
	return JobBinding{binder: binder}, nil
}

func (binding JobBinding) Snapshot(
	ctx context.Context,
	scope platformtx.JobTenantScope,
	operationID uuid.UUID,
) (JobSnapshot, error) {
	if ctx == nil || ctx.Err() != nil || operationID == uuid.Nil ||
		!platformtx.ContextOwnsJobScope(ctx, scope) || platformtx.NetworkAllowed(ctx) {
		return JobSnapshot{}, catalogapp.ErrForbidden
	}
	reader, err := platformtx.BindJob(scope, binding.binder)
	if err != nil || reader == nil || reader.queryRow == nil {
		return JobSnapshot{}, catalogapp.ErrUnavailable
	}
	operation, _, err := scanOperation(reader.queryRow(ctx, getOperationSQL, operationID))
	if err != nil {
		return JobSnapshot{}, catalogapp.ErrUnavailable
	}
	var actorID uuid.UUID
	if err = reader.queryRow(ctx, getOperationActorSQL, operationID).Scan(&actorID); err != nil ||
		actorID == uuid.Nil {
		return JobSnapshot{}, catalogapp.ErrUnavailable
	}
	return JobSnapshot{Operation: operation, ActorID: actorID}, nil
}

func (binder *Binder) Bind(
	ctx context.Context, scope platformtx.TenantScope, access catalogapp.ScopeAccess,
) (catalogapp.Store, catalogapp.ScopeIdentity, error) {
	mode, ok := accessMode(access)
	identity := catalogapp.ScopeIdentity{ActorID: scope.ActorID(), TenantID: scope.TenantID()}
	if binder == nil || !ok || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		!scopeUsable(ctx, scope, mode) {
		return nil, catalogapp.ScopeIdentity{}, unavailable()
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{tenantID: identity.TenantID, actorID: identity.ActorID,
				command: mode == platformtx.Command,
				queryRow: func(callCtx context.Context, statement string, arguments ...any) row {
					return database.QueryRow(callCtx, statement, arguments...)
				}, query: func(callCtx context.Context, statement string, arguments ...any) (rows, error) {
					return database.Query(callCtx, statement, arguments...)
				}, execute: func(callCtx context.Context, statement string, arguments ...any) (int64, error) {
					result, executeErr := database.Execute(callCtx, statement, arguments...)
					return result.RowsAffected(), executeErr
				}, currentSchema: database.CurrentSchema, notFound: platformpostgres.RowNotFound,
				allowed: func(callCtx context.Context) bool { return scopeUsable(callCtx, scope, mode) }}
		})
	if err != nil {
		return nil, catalogapp.ScopeIdentity{}, unavailable()
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, catalogapp.ScopeIdentity{}, unavailable()
	}
	return bound, identity, nil
}

func accessMode(access catalogapp.ScopeAccess) (platformtx.ScopeMode, bool) {
	switch access {
	case catalogapp.ScopeCommand:
		return platformtx.Command, true
	case catalogapp.ScopeSnapshot:
		return platformtx.SnapshotQuery, true
	default:
		return "", false
	}
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

var _ catalogapp.StoreBinder = (*Binder)(nil)
