package async

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validDatabaseEffectEvent() DatabaseJobEventInput {
	payload := json.RawMessage(`{"value":1}`)
	digest := sha256.Sum256(payload)
	return DatabaseJobEventInput{
		EventID:   uuid.MustParse("019c6a7e-7c00-7000-8000-000000000010"),
		EventType: "billing.subscription.changed", EventVersion: 1,
		AggregateType: "billing.subscription", AggregateID: uuid.New(),
		AggregateVersion: 2, EventSequence: 2,
		OccurredAt: time.Date(2026, 7, 17, 1, 2, 3, 456000000, time.UTC),
		ActorID:    uuid.New(), CorrelationID: "billing:job:1", CausationID: uuid.New(),
		IdempotencyKey:               "billing-event-1",
		BusinessEffectIdempotencyKey: "billing-subscription-1:2",
		PayloadSchemaSHA256:          strings.Repeat("a", 64),
		PayloadSHA256:                hex.EncodeToString(digest[:]), Payload: payload,
	}
}

func validDatabaseEffectCancellation() DatabaseJobCancellationInput {
	return DatabaseJobCancellationInput{
		JobID:       uuid.MustParse("019c6a7e-7c00-7000-8000-000000000020"),
		RequestedAt: time.Date(2026, 7, 17, 1, 2, 3, 456000000, time.UTC),
		Code:        "BILLING_BOUNDARY_CANCELLED", Digest: strings.Repeat("b", 64),
	}
}

func TestDatabaseJobEffectsAreBoundedAndDeeplySealed(t *testing.T) {
	event := validDatabaseEffectEvent()
	cancellation := validDatabaseEffectCancellation()
	outcome, err := CompleteDatabaseJobWithEffects(DatabaseJobEffects{
		Cancellations: []DatabaseJobCancellationInput{cancellation},
		Events:        []DatabaseJobEventInput{event},
	})
	if err != nil || !outcome.Valid() {
		t.Fatalf("complete database job effects = %+v, %v", outcome, err)
	}
	event.Payload[2] = 'X'
	first := outcome.Events()
	if string(first[0].Payload) != `{"value":1}` {
		t.Fatalf("sealed payload = %s", first[0].Payload)
	}
	first[0].Payload[2] = 'Y'
	if got := string(outcome.Events()[0].Payload); got != `{"value":1}` {
		t.Fatalf("returned payload alias changed outcome: %s", got)
	}
	if got := outcome.Cancellations(); len(got) != 1 || got[0] != cancellation {
		t.Fatalf("sealed cancellations = %#v", got)
	}
}

func TestDatabaseJobEffectsRejectInvalidEventWire(t *testing.T) {
	tests := map[string]func(*DatabaseJobEventInput){
		"non canonical payload": func(value *DatabaseJobEventInput) {
			value.Payload = json.RawMessage(`{"value": 1}`)
		},
		"invalid type": func(value *DatabaseJobEventInput) {
			value.EventType = "Billing.Changed"
		},
		"invalid correlation": func(value *DatabaseJobEventInput) {
			value.CorrelationID = " billing"
		},
		"padded key": func(value *DatabaseJobEventInput) {
			value.IdempotencyKey = " padded"
		},
		"sub microsecond time": func(value *DatabaseJobEventInput) {
			value.OccurredAt = value.OccurredAt.Add(time.Nanosecond)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			event := validDatabaseEffectEvent()
			mutate(&event)
			if _, err := CompleteDatabaseJobWithEffects(
				DatabaseJobEffects{Events: []DatabaseJobEventInput{event}},
			); err == nil {
				t.Fatal("invalid event was accepted")
			}
		})
	}
}

func TestDatabaseJobEffectsRejectAmbiguousOrUnboundedSet(t *testing.T) {
	cancellation := validDatabaseEffectCancellation()
	successor := DatabaseJobSuccessorInput{
		JobID: cancellation.JobID, RunAt: cancellation.RequestedAt,
		PayloadReference: "urn:harpy:billing:boundary:1",
		PayloadSHA256:    strings.Repeat("c", 64), DedupeKey: "billing:boundary:1",
	}
	if _, err := CompleteDatabaseJobWithEffects(DatabaseJobEffects{
		Successor:     &successor,
		Cancellations: []DatabaseJobCancellationInput{cancellation},
	}); err == nil {
		t.Fatal("successor cancellation ambiguity was accepted")
	}
	cancellations := make([]DatabaseJobCancellationInput, maximumDatabaseJobEffects+1)
	for index := range cancellations {
		cancellations[index] = cancellation
		cancellations[index].JobID = uuid.Must(uuid.NewV7())
	}
	if _, err := CompleteDatabaseJobWithEffects(DatabaseJobEffects{
		Cancellations: cancellations,
	}); err == nil {
		t.Fatal("unbounded effect set was accepted")
	}
}
