package analytics

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type publicationTransitionPayload struct {
	ReleaseID         string `json:"release_id"`
	ResourceKind      string `json:"resource_kind"`
	ResourceVersionID string `json:"resource_version_id"`
	Audience          string `json:"audience"`
	Generation        uint64 `json:"generation"`
}

func (effects *AsyncEffects) Activated(
	ctx context.Context, scope platformtx.TenantScope, effect publicationapp.TransitionEffect,
) error {
	return effects.publicationTransition(ctx, scope, effect,
		asyncmanifestv1.AnalyticsPublicationActivatedEventType,
		asyncmanifestv1.AnalyticsPublicationActivatedSchemaSHA256, "active")
}

func (effects *AsyncEffects) Revoked(
	ctx context.Context, scope platformtx.TenantScope, effect publicationapp.TransitionEffect,
) error {
	return effects.publicationTransition(ctx, scope, effect,
		asyncmanifestv1.AnalyticsPublicationRevokedEventType,
		asyncmanifestv1.AnalyticsPublicationRevokedSchemaSHA256, "revoked")
}

func (effects *AsyncEffects) publicationTransition(
	ctx context.Context, scope platformtx.TenantScope,
	effect publicationapp.TransitionEffect, eventType, schema, state string,
) error {
	writer, authority, err := effects.bind(scope)
	if err != nil || effect.TenantID != authority.ScopeID() ||
		effect.ActorID != scope.ActorID() || ctx == nil || ctx.Err() != nil {
		return publicationapp.ErrUnavailable
	}
	envelope, err := effects.event(authority, analyticsEventInput{
		EventID: effect.OperationID, ActorID: effect.ActorID,
		EventType: eventType, Schema: schema, AggregateType: "analytics.publication",
		AggregateID: effect.ReleaseID, AggregateVersion: effect.Generation,
		EventSequence: effect.Generation, OccurredAt: effect.OccurredAt,
		Payload: publicationTransitionPayload{
			ReleaseID: effect.ReleaseID.String(), ResourceKind: string(effect.ResourceKind),
			ResourceVersionID: effect.ResourceVersionID.String(),
			Audience:          string(effect.Audience), Generation: effect.Generation,
		},
	})
	if err != nil {
		return publicationapp.ErrUnavailable
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || !result.Created || result.ID != effect.OperationID.String() {
		return publicationapp.ErrUnavailable
	}
	reference := fmt.Sprintf("analytics-publication-repair/%s/%d/%s",
		effect.ReleaseID, effect.Generation, state)
	if err := effects.enqueueJob(ctx, writer, authority,
		asyncmanifestv1.AnalyticsPublicationRepairJobKind, reference,
		fmt.Sprintf("analytics.publication.repair:%s:%d", effect.ReleaseID, effect.Generation),
		effect.OccurredAt); err != nil {
		return publicationapp.ErrUnavailable
	}
	return nil
}

var _ publicationapp.Effects = (*AsyncEffects)(nil)
