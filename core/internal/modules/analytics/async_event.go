package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type analyticsEventInput struct {
	EventID, ActorID                uuid.UUID
	EventType, Schema               string
	AggregateType                   string
	AggregateID                     uuid.UUID
	AggregateVersion, EventSequence uint64
	OccurredAt                      time.Time
	Payload                         any
}

func (effects *AsyncEffects) event(
	authority platformtx.OwnerAuthority, input analyticsEventInput,
) (eventsv1.Envelope, error) {
	if input.AggregateVersion == 0 || input.AggregateVersion > math.MaxInt64 ||
		input.EventSequence == 0 || input.EventSequence > math.MaxInt64 {
		return eventsv1.Envelope{}, fmt.Errorf("Analytics event version is invalid")
	}
	wire, err := json.Marshal(input.Payload)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		wire, eventsv1.MaximumCanonicalPayloadBytes)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	digest := sha256.Sum256(canonical)
	tenantID := authority.ScopeID().String()
	version := int64(input.AggregateVersion)
	sequence := int64(input.EventSequence)
	value := eventsv1.Envelope{
		EventID: input.EventID.String(), EventType: input.EventType, EventVersion: 1,
		TenantID: &tenantID, ProducerAuthority: analyticsEventAuthority(authority),
		AggregateType: input.AggregateType, AggregateID: input.AggregateID.String(),
		AggregateVersion: version, EventSequence: sequence,
		OccurredAt: input.OccurredAt.Format(time.RFC3339Nano),
		ActorID:    input.ActorID.String(), CorrelationID: input.EventID.String(),
		CausationID: input.EventID.String(), IdempotencyKey: input.EventID.String(),
		BusinessEffectIdempotencyKey: fmt.Sprintf(
			"%s:%s:%d", input.EventType, input.AggregateID, sequence),
		PayloadSchemaSHA256: input.Schema,
		PayloadSHA256:       hex.EncodeToString(digest[:]), Payload: canonical,
	}
	_, producers, err := effects.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	return eventsv1.NewEnvelope(value, producers)
}

func analyticsEventAuthority(authority platformtx.OwnerAuthority) eventsv1.Authority {
	return eventsv1.Authority{ScopeKind: eventsv1.ScopeTenant,
		ScopeID: authority.ScopeID().String(), Domain: authority.Domain(),
		CutoverGeneration: authority.Generation(), WriterOrigin: eventsv1.WriterV2,
		GroupID: authority.GroupID(), GroupGeneration: authority.GroupGeneration()}
}
