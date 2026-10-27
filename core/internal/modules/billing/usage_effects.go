package billing

import (
	"context"
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	usagepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/usage"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const usageThresholdSchemaSHA256 = "69e78f61bc3db12a3c72883fab952d6d87a29f9ceafec54bb3772fa03647d591"

type usageEffects struct {
	tenantID  uuid.UUID
	actorID   uuid.UUID
	authority platformtx.OwnerAuthority
	producers eventsv1.ProducerCatalog
	writer    asyncpostgres.Writer
}

func newUsageEffects(
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
	catalogs platformasync.WriteCatalogs,
) (*usageEffects, error) {
	_, producers, err := catalogs.Snapshot()
	if err != nil || !authority.OwnsScope(scope) || authority.Domain() != "billing" {
		return nil, usageapp.ErrConfiguration
	}
	writer, err := asyncpostgres.Bind(scope, authority, catalogs)
	if err != nil {
		return nil, fmt.Errorf("bind billing usage async writer: %w", err)
	}
	return &usageEffects{
		tenantID: scope.TenantID(), actorID: scope.ActorID(), authority: authority,
		producers: producers, writer: writer,
	}, nil
}

func (effects *usageEffects) PublishThreshold(
	ctx context.Context,
	event usageapp.ThresholdEventIntent,
) error {
	if effects == nil || !event.Valid() {
		return usageapp.ErrConfiguration
	}
	payload, digest, err := encodeUsageThresholdEvent(event)
	if err != nil {
		return err
	}
	tenantID, causationID := effects.tenantID.String(), event.CausationID.String()
	envelope, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: event.ID.String(), EventType: usageapp.ThresholdReachedEventType,
		EventVersion: 1, TenantID: &tenantID,
		ProducerAuthority: billingEventAuthority(effects.authority),
		AggregateType:     "billing.usage", AggregateID: event.MeterPeriodID.String(),
		AggregateVersion: event.UsageRevision, EventSequence: event.UsageRevision,
		OccurredAt: event.OccurredAt.Format(time.RFC3339Nano),
		ActorID:    effects.actorID.String(), CorrelationID: causationID,
		CausationID: causationID, IdempotencyKey: event.ID.String(),
		BusinessEffectIdempotencyKey: fmt.Sprintf(
			"billing.usage.threshold_reached:%s:%d",
			event.MeterPeriodID, event.UsageRevision,
		),
		PayloadSchemaSHA256: usageThresholdSchemaSHA256,
		PayloadSHA256:       digest, Payload: payload,
	}, effects.producers)
	if err != nil {
		return err
	}
	result, err := effects.writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != event.ID.String() {
		return fmt.Errorf("enqueue billing usage event %s: %w", event.ID, err)
	}
	return nil
}

var _ usagepostgres.Effects = (*usageEffects)(nil)
