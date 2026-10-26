package postgres

import (
	"context"
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func (store *boundDatabaseJobStore) enqueueEvent(
	ctx context.Context,
	claim JobClaim,
	input platformasync.DatabaseJobEventInput,
) error {
	tenantID := claim.tenantID.String()
	value, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: input.EventID.String(), EventType: input.EventType,
		EventVersion: input.EventVersion, TenantID: &tenantID,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: tenantID,
			Domain:            claim.authority.domain,
			CutoverGeneration: claim.authority.generation,
			WriterOrigin:      eventsv1.WriterV2, GroupID: claim.authority.groupID,
			GroupGeneration: claim.authority.groupGeneration,
		},
		AggregateType: input.AggregateType, AggregateID: input.AggregateID.String(),
		AggregateVersion: input.AggregateVersion, EventSequence: input.EventSequence,
		OccurredAt: input.OccurredAt.Format(time.RFC3339Nano),
		ActorID:    input.ActorID.String(), CorrelationID: input.CorrelationID,
		CausationID: input.CausationID.String(), IdempotencyKey: input.IdempotencyKey,
		BusinessEffectIdempotencyKey: input.BusinessEffectIdempotencyKey,
		PayloadSchemaSHA256:          input.PayloadSchemaSHA256,
		PayloadSHA256:                input.PayloadSHA256, Payload: input.Payload,
	}, store.producers)
	if err != nil {
		return fmt.Errorf("validate database job event: %w", err)
	}
	repository := &repository{
		database: store.database,
		authority: authorityStamp{
			scopeID: claim.tenantID, groupID: claim.authority.groupID,
			groupGeneration: claim.authority.groupGeneration,
			domain:          claim.authority.domain, generation: claim.authority.generation,
		},
		producers: store.producers,
	}
	result, err := repository.EnqueueEvent(ctx, value)
	if err != nil || result.ID != input.EventID.String() {
		return fmt.Errorf("enqueue database job event: %w", err)
	}
	return nil
}
