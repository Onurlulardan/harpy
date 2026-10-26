package work

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (publisher *eventPublisher) envelope(authority platformtx.OwnerAuthority,
	local ports.Scope, intent ports.EventIntent,
) (eventsv1.Envelope, error) {
	var result any
	if len(intent.Payload) == 0 {
		result = nil
	} else if err := json.Unmarshal(intent.Payload, &result); err != nil {
		return eventsv1.Envelope{}, err
	}
	wire, err := json.Marshal(map[string]any{"operation": intent.Type, "result": result})
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	payload, err := canonicaljsonv1.CanonicalizeObject(wire, eventsv1.MaximumCanonicalPayloadBytes)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	digest := sha256.Sum256(payload)
	tenantID := local.TenantID().String()
	value := eventsv1.Envelope{
		EventID: intent.EventID.String(), EventType: intent.Type, EventVersion: 1,
		TenantID: &tenantID, ProducerAuthority: workEventAuthority(authority),
		AggregateType: intent.AggregateKind, AggregateID: intent.AggregateID.String(),
		AggregateVersion: intent.AggregateVersion, EventSequence: intent.EventSequence,
		OccurredAt: intent.OccurredAt.UTC().Format(time.RFC3339Nano),
		ActorID:    local.ActorID().String(), CorrelationID: intent.EventID.String(),
		CausationID: intent.EventID.String(), IdempotencyKey: intent.IdempotencyKey,
		BusinessEffectIdempotencyKey: fmt.Sprintf("%s:%s:%d", intent.Type,
			intent.AggregateID, intent.EventSequence),
		PayloadSchemaSHA256: asyncmanifestv1.WorkMutationSchemaSHA256,
		PayloadSHA256:       hex.EncodeToString(digest[:]), Payload: payload,
	}
	_, producers, err := publisher.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	return eventsv1.NewEnvelope(value, producers)
}

func workEventAuthority(authority platformtx.OwnerAuthority) eventsv1.Authority {
	return eventsv1.Authority{ScopeKind: eventsv1.ScopeTenant,
		ScopeID: authority.ScopeID().String(), Domain: authority.Domain(),
		CutoverGeneration: authority.Generation(), WriterOrigin: eventsv1.WriterV2,
		GroupID: authority.GroupID(), GroupGeneration: authority.GroupGeneration()}
}

func workMutationTypes() []string {
	policies := asyncmanifestv1.WorkProducerPolicies()
	result := make([]string, 0, len(policies)-1)
	for _, policy := range policies {
		if policy.EventType != asyncmanifestv1.WorkReminderDueEventType {
			result = append(result, policy.EventType)
		}
	}
	return result
}
