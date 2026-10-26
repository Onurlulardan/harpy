package billing

import (
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func (effects *subscriptionEffects) subscriptionEnvelope(
	intent subscriptionapp.OutboxIntent,
	snapshot subscriptionapp.SubscriptionSnapshot,
) (eventsv1.Envelope, error) {
	payload, digest, schema, aggregate, effectKey, err := subscriptionEventDocument(
		intent, snapshot,
	)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	tenantID := effects.tenantID.String()
	causationID := intent.CausationID.String()
	value := eventsv1.Envelope{
		EventID: intent.ID.String(), EventType: intent.EventType,
		EventVersion: intent.EventVersion, TenantID: &tenantID,
		ProducerAuthority: billingEventAuthority(effects.authority),
		AggregateType:     aggregate, AggregateID: snapshot.ID.String(),
		AggregateVersion: intent.AggregateRevision,
		EventSequence:    intent.AggregateRevision,
		OccurredAt:       intent.OccurredAt.Format(time.RFC3339Nano),
		ActorID:          intent.ActorID.String(), CorrelationID: causationID,
		CausationID: causationID, IdempotencyKey: intent.ID.String(),
		BusinessEffectIdempotencyKey: effectKey,
		PayloadSchemaSHA256:          schema,
		PayloadSHA256:                digest, Payload: payload,
	}
	return eventsv1.NewEnvelope(value, effects.producers)
}

func subscriptionEventDocument(
	intent subscriptionapp.OutboxIntent,
	snapshot subscriptionapp.SubscriptionSnapshot,
) ([]byte, string, string, string, string, error) {
	var payload []byte
	var digest, schema, aggregate, effectPrefix string
	var err error
	switch intent.EventType {
	case "billing.subscription.changed":
		payload, digest, err = encodeSubscriptionEvent(intent, snapshot)
		schema, aggregate = subscriptionChangeSchemaSHA256, "billing.subscription"
		effectPrefix = "billing.subscription.changed"
	case "billing.entitlements.changed":
		payload, digest, err = encodeEntitlementEvent(snapshot)
		schema, aggregate = entitlementChangeSchemaSHA256, "billing.entitlements"
		effectPrefix = "billing.entitlements.changed"
	default:
		err = fmt.Errorf("unsupported billing subscription event %q", intent.EventType)
	}
	effectKey := fmt.Sprintf("%s:%s:%d", effectPrefix, snapshot.ID, snapshot.Revision)
	return payload, digest, schema, aggregate, effectKey, err
}
