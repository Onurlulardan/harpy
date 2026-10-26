package analytics

import (
	"context"
	"encoding/hex"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type dashboardPublishedPayload struct {
	DashboardID        string `json:"dashboard_id"`
	DashboardVersionID string `json:"dashboard_version_id"`
	RootRevision       uint64 `json:"root_revision"`
	ManifestSHA256     string `json:"manifest_sha256"`
}

func (effects *AsyncEffects) DashboardPublished(
	ctx context.Context, scope platformtx.TenantScope, effect dashboardapp.PublishEffect,
) error {
	writer, authority, err := effects.bind(scope)
	if err != nil || effect.TenantID != authority.ScopeID() ||
		effect.ActorID != scope.ActorID() || ctx == nil || ctx.Err() != nil {
		return dashboardapp.ErrUnavailable
	}
	manifest := hex.EncodeToString(effect.ManifestSHA256[:])
	envelope, err := effects.event(authority, analyticsEventInput{
		EventID: effect.OperationID, ActorID: effect.ActorID,
		EventType:     asyncmanifestv1.AnalyticsDashboardPublishedEventType,
		Schema:        asyncmanifestv1.AnalyticsDashboardPublishedSchemaSHA256,
		AggregateType: "analytics.dashboard", AggregateID: effect.DashboardID,
		AggregateVersion: effect.RootRevision, EventSequence: effect.VersionNumber,
		OccurredAt: effect.OccurredAt,
		Payload: dashboardPublishedPayload{DashboardID: effect.DashboardID.String(),
			DashboardVersionID: effect.DashboardVersionID.String(),
			RootRevision:       effect.RootRevision, ManifestSHA256: manifest},
	})
	if err != nil {
		return dashboardapp.ErrUnavailable
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || !result.Created || result.ID != effect.OperationID.String() {
		return dashboardapp.ErrUnavailable
	}
	return nil
}

var _ dashboardapp.Effects = (*AsyncEffects)(nil)
