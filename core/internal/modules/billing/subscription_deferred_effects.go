package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	subscriptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/subscription"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type deferredSubscriptionEffects struct {
	tenantID      uuid.UUID
	successor     *platformasync.DatabaseJobSuccessorInput
	cancellations []platformasync.DatabaseJobCancellationInput
	events        []platformasync.DatabaseJobEventInput
}

func newDeferredSubscriptionEffects(tenantID uuid.UUID) *deferredSubscriptionEffects {
	return &deferredSubscriptionEffects{tenantID: tenantID}
}

func (effects *deferredSubscriptionEffects) EnqueueBoundary(
	_ context.Context,
	intent subscriptionapp.JobIntent,
) (subscriptionpostgres.BoundaryArtifact, error) {
	if effects == nil || effects.tenantID == uuid.Nil || effects.successor != nil {
		return subscriptionpostgres.BoundaryArtifact{}, subscriptionapp.ErrSnapshot
	}
	artifact, err := subscriptionBoundaryArtifact(effects.tenantID, intent)
	if err != nil {
		return subscriptionpostgres.BoundaryArtifact{}, err
	}
	effects.successor = &platformasync.DatabaseJobSuccessorInput{
		JobID: intent.ID, RunAt: intent.RunAt,
		PayloadReference: artifact.PayloadReference,
		PayloadSHA256:    hex.EncodeToString(artifact.PayloadSHA256[:]),
		DedupeKey:        "billing.subscription.boundary:" + intent.EffectID.String(),
		Priority:         100,
	}
	return artifact, nil
}

func (effects *deferredSubscriptionEffects) CancelBoundary(
	_ context.Context,
	cancellation subscriptionapp.JobCancellation,
) error {
	if effects == nil {
		return subscriptionapp.ErrConfiguration
	}
	digest := sha256.Sum256([]byte(cancellation.Reason))
	effects.cancellations = append(
		effects.cancellations,
		platformasync.DatabaseJobCancellationInput{
			JobID: cancellation.ID, RequestedAt: cancellation.CancelledAt,
			Code: "BILLING_BOUNDARY_CANCELLED", Digest: hex.EncodeToString(digest[:]),
		},
	)
	return nil
}

func (effects *deferredSubscriptionEffects) PublishSubscription(
	_ context.Context,
	intent subscriptionapp.OutboxIntent,
	snapshot subscriptionapp.SubscriptionSnapshot,
) error {
	if effects == nil {
		return subscriptionapp.ErrConfiguration
	}
	payload, digest, schema, aggregate, effectKey, err := subscriptionEventDocument(
		intent, snapshot,
	)
	if err != nil {
		return err
	}
	causationID := intent.CausationID.String()
	effects.events = append(effects.events, platformasync.DatabaseJobEventInput{
		EventID: intent.ID, EventType: intent.EventType,
		EventVersion: intent.EventVersion, AggregateType: aggregate,
		AggregateID: snapshot.ID, AggregateVersion: intent.AggregateRevision,
		EventSequence: intent.AggregateRevision, OccurredAt: intent.OccurredAt,
		ActorID: intent.ActorID, CausationID: intent.CausationID,
		CorrelationID: causationID, IdempotencyKey: intent.ID.String(),
		BusinessEffectIdempotencyKey: effectKey,
		PayloadSchemaSHA256:          schema,
		PayloadSHA256:                digest, Payload: payload,
	})
	return nil
}

func (effects *deferredSubscriptionEffects) outcome() (
	platformasync.DatabaseJobOutcome,
	error,
) {
	if effects == nil {
		return platformasync.DatabaseJobOutcome{}, subscriptionapp.ErrConfiguration
	}
	return platformasync.CompleteDatabaseJobWithEffects(platformasync.DatabaseJobEffects{
		Successor: effects.successor, Cancellations: effects.cancellations,
		Events: effects.events,
	})
}

var _ subscriptionpostgres.Effects = (*deferredSubscriptionEffects)(nil)
