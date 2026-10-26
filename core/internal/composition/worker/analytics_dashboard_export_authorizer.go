package worker

import (
	"context"
	"fmt"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type analyticsDashboardExportAuthorizer struct {
	pool        *platformpostgres.Pool
	coordinator *platformtx.Coordinator
	iam         *iammodule.WorkerAuthorizationRuntime
}

func (authorizer analyticsDashboardExportAuthorizer) Authorize(
	ctx context.Context, request analyticsmodule.DashboardExportAuthorizationRequest,
) error {
	if authorizer.pool == nil || authorizer.coordinator == nil || authorizer.iam == nil ||
		ctx == nil || ctx.Err() != nil || request.TenantID == [16]byte{} ||
		request.ActorID == [16]byte{} || request.DashboardID.Version() != 7 ||
		request.TokenVersion < 1 || request.AuthorizationVersion < 1 ||
		request.AuthorizationClock < 1 ||
		request.Permission != string(analyticsmodule.PermissionDashboardExport) {
		return fmt.Errorf("Analytics dashboard export authorization is invalid")
	}
	_, generation, err := authorizer.pool.ResolveTenantAnalyticsGenerations(
		ctx, request.TenantID)
	if err != nil || generation < 1 {
		return fmt.Errorf("resolve Analytics dashboard export authority: %w", err)
	}
	actor, err := actorcontext.Issue(request.ActorID, request.TenantID)
	if err != nil {
		return fmt.Errorf("issue Analytics dashboard export actor: %w", err)
	}
	operation, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		"analytics.dashboard.export.worker.authorize",
		map[string]int64{"analytics": generation},
	)
	if err != nil {
		return fmt.Errorf("issue Analytics dashboard export authority: %w", err)
	}
	return authorizer.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: operation, Mode: platformtx.SnapshotQuery,
	}, func(scoped context.Context, scope platformtx.TenantScope) error {
		for _, target := range [][2]string{
			{"analytics.dashboard", "export"},
			{"analytics.query", "execute"},
		} {
			allowed, _, tokenVersion, authorizationVersion, clock, authorizeErr :=
				authorizer.iam.AuthorizePinned(scoped, scope, target[0], target[1],
					request.AuthorizationVersion)
			if authorizeErr != nil || !allowed || tokenVersion != request.TokenVersion ||
				authorizationVersion != request.AuthorizationVersion ||
				clock != request.AuthorizationClock {
				return fmt.Errorf("Analytics dashboard export permission is stale")
			}
		}
		return nil
	})
}

var _ analyticsmodule.DashboardExportWorkerAuthorizer = analyticsDashboardExportAuthorizer{}
