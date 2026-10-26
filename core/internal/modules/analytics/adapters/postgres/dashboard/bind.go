package dashboard

import (
	"context"
	_ "embed"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context, scope platformtx.TenantScope,
	access dashboardapp.ScopeAccess,
) (dashboardapp.Store, dashboardapp.ScopeIdentity, error) {
	mode, ok := accessMode(access)
	identity := dashboardapp.ScopeIdentity{ActorID: scope.ActorID(), TenantID: scope.TenantID()}
	if binder == nil || !ok || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		!scopeUsable(ctx, scope, mode) {
		return nil, dashboardapp.ScopeIdentity{}, dashboardapp.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{actorID: identity.ActorID, command: mode == platformtx.Command,
				queryRow: func(ctx context.Context, sql string, args ...any) row {
					return database.QueryRow(ctx, sql, args...)
				}, query: func(ctx context.Context, sql string, args ...any) (rows, error) {
					return database.Query(ctx, sql, args...)
				}, allowed: func(ctx context.Context) bool { return scopeUsable(ctx, scope, mode) },
				notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, dashboardapp.ScopeIdentity{}, dashboardapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, dashboardapp.ScopeIdentity{}, dashboardapp.ErrUnavailable
	}
	return bound, identity, nil
}

func accessMode(access dashboardapp.ScopeAccess) (platformtx.ScopeMode, bool) {
	if access == dashboardapp.ScopeCommand {
		return platformtx.Command, true
	}
	if access == dashboardapp.ScopeSnapshot {
		return platformtx.SnapshotQuery, true
	}
	return "", false
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

var _ dashboardapp.StoreBinder = (*Binder)(nil)

//go:embed sql/tenant/get_version.sql
var getVersionSQL string

type PublicSource struct{}

func NewPublicSource() *PublicSource { return &PublicSource{} }

func (source *PublicSource) LoadPublishedVersion(
	ctx context.Context, scope platformtx.TenantScope,
	dashboardID, versionID uuid.UUID,
) (dashboardapp.Snapshot, error) {
	if source == nil || !validUUIDv7(dashboardID) || !validUUIDv7(versionID) ||
		!scopeUsable(ctx, scope, platformtx.SnapshotQuery) ||
		platformtx.RequireOwner(scope, storageOwner) != nil ||
		scope.OperationID() != "analytics.publication.execute" {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	binder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) queryRow {
			return func(call context.Context, statement string, arguments ...any) row {
				return database.QueryRow(call, statement, arguments...)
			}
		})
	if err != nil {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	result, err := scanSnapshot(bound(ctx, getVersionSQL, dashboardID, versionID),
		platformpostgres.RowNotFound)
	if err != nil || result.DashboardID != dashboardID ||
		result.VersionID != versionID || !result.Published ||
		result.Lifecycle != analyticsdashboard.LifecycleActive {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	return result, nil
}
