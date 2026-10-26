package analytics

import (
	"context"
	"encoding/hex"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type queryPublishedPayload struct {
	QueryID        string `json:"query_id"`
	QueryVersionID string `json:"query_version_id"`
	RootRevision   uint64 `json:"root_revision"`
	ManifestSHA256 string `json:"manifest_sha256"`
}

func (effects *AsyncEffects) QueryPublished(
	ctx context.Context, scope platformtx.TenantScope, effect queryapp.PublishEffect,
) error {
	writer, authority, err := effects.bind(scope)
	if err != nil || effect.TenantID != authority.ScopeID() ||
		effect.ActorID != scope.ActorID() || ctx == nil || ctx.Err() != nil {
		return queryapp.ErrUnavailable
	}
	manifest := hex.EncodeToString(effect.ManifestSHA256[:])
	envelope, err := effects.event(authority, analyticsEventInput{
		EventID: effect.OperationID, ActorID: effect.ActorID,
		EventType:     asyncmanifestv1.AnalyticsQueryPublishedEventType,
		Schema:        asyncmanifestv1.AnalyticsQueryPublishedSchemaSHA256,
		AggregateType: "analytics.query", AggregateID: effect.QueryID,
		AggregateVersion: effect.RootRevision, EventSequence: effect.VersionNumber,
		OccurredAt: effect.OccurredAt,
		Payload: queryPublishedPayload{QueryID: effect.QueryID.String(),
			QueryVersionID: effect.QueryVersionID.String(),
			RootRevision:   effect.RootRevision, ManifestSHA256: manifest},
	})
	if err != nil {
		return queryapp.ErrUnavailable
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || !result.Created || result.ID != effect.OperationID.String() {
		return queryapp.ErrUnavailable
	}
	if effect.Revalidated {
		reference := fmt.Sprintf("analytics-query-revalidate/%s/%s/%s",
			effect.QueryID, effect.QueryVersionID, manifest)
		if err := effects.enqueueJob(ctx, writer, authority,
			asyncmanifestv1.AnalyticsQueryRevalidateJobKind, reference,
			"analytics.query.revalidate:"+effect.QueryVersionID.String(), effect.OccurredAt); err != nil {
			return queryapp.ErrUnavailable
		}
	}
	reference := fmt.Sprintf("analytics-cache-warm/%s/%s/%s",
		effect.QueryID, effect.QueryVersionID, manifest)
	if err := effects.enqueueJob(ctx, writer, authority,
		asyncmanifestv1.AnalyticsCacheWarmJobKind, reference,
		"analytics.cache.warm:"+effect.QueryVersionID.String(), effect.OccurredAt); err != nil {
		return queryapp.ErrUnavailable
	}
	return nil
}

var _ queryapp.Effects = (*AsyncEffects)(nil)
