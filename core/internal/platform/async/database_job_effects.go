package async

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

const maximumDatabaseJobEffects = 16

var databaseJobErrorCode = regexp.MustCompile(`^[A-Z][A-Z0-9_]{2,63}$`)
var databaseJobEventType = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
var databaseJobCorrelation = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,63}$`)

type DatabaseJobCancellationInput struct {
	JobID       uuid.UUID
	RequestedAt time.Time
	Code        string
	Digest      string
}

type DatabaseJobEventInput struct {
	EventID, AggregateID, ActorID, CausationID uuid.UUID
	EventType, AggregateType                   string
	EventVersion                               int
	AggregateVersion, EventSequence            int64
	OccurredAt                                 time.Time
	CorrelationID, IdempotencyKey              string
	BusinessEffectIdempotencyKey               string
	PayloadSchemaSHA256, PayloadSHA256         string
	Payload                                    json.RawMessage
}

type DatabaseJobEffects struct {
	Successor     *DatabaseJobSuccessorInput
	Children      []DatabaseJobChildInput
	Cancellations []DatabaseJobCancellationInput
	Events        []DatabaseJobEventInput
}

func CompleteDatabaseJobWithEffects(
	effects DatabaseJobEffects,
) (DatabaseJobOutcome, error) {
	effectCount := len(effects.Children) + len(effects.Cancellations) + len(effects.Events)
	if effects.Successor != nil {
		effectCount++
	}
	if effectCount > maximumDatabaseJobEffects ||
		!validDatabaseJobChildren(effects.Children) ||
		!validDatabaseJobCancellations(effects.Cancellations) ||
		!validDatabaseJobEvents(effects.Events) ||
		databaseJobChildConflicts(
			effects.Children, effects.Successor, effects.Cancellations,
		) ||
		effects.Successor != nil && (!validDatabaseJobSuccessor(*effects.Successor) ||
			cancelsDatabaseJob(effects.Cancellations, effects.Successor.JobID)) {
		return DatabaseJobOutcome{}, ErrDatabaseJobOutcomeInvalid
	}
	outcome := DatabaseJobOutcome{
		children:      append([]DatabaseJobChildInput(nil), effects.Children...),
		cancellations: append([]DatabaseJobCancellationInput(nil), effects.Cancellations...),
		events:        make([]DatabaseJobEventInput, len(effects.Events)), valid: true,
	}
	for index := range effects.Events {
		outcome.events[index] = effects.Events[index].clone()
	}
	if effects.Successor != nil {
		outcome.successor, outcome.hasNext = *effects.Successor, true
	}
	return outcome, nil
}

func cancelsDatabaseJob(
	cancellations []DatabaseJobCancellationInput,
	jobID uuid.UUID,
) bool {
	for _, cancellation := range cancellations {
		if cancellation.JobID == jobID {
			return true
		}
	}
	return false
}

func validDatabaseJobCancellations(values []DatabaseJobCancellationInput) bool {
	seen := map[uuid.UUID]bool{}
	for _, value := range values {
		if value.JobID == uuid.Nil || value.JobID.Version() != uuid.Version(7) ||
			value.RequestedAt.IsZero() || value.RequestedAt.Location() != time.UTC ||
			value.RequestedAt.Nanosecond()%int(time.Microsecond) != 0 ||
			!databaseJobErrorCode.MatchString(value.Code) ||
			!databaseJobDigest.MatchString(value.Digest) || seen[value.JobID] {
			return false
		}
		seen[value.JobID] = true
	}
	return true
}

func validDatabaseJobEvents(values []DatabaseJobEventInput) bool {
	seen := map[uuid.UUID]bool{}
	for _, value := range values {
		canonical, err := canonicaljsonv1.CanonicalizeObject(
			value.Payload, eventsv1.MaximumCanonicalPayloadBytes,
		)
		digest := sha256.Sum256(canonical)
		if err != nil || !bytes.Equal(canonical, value.Payload) ||
			value.EventID == uuid.Nil || value.EventID.Version() != uuid.Version(7) ||
			value.AggregateID == uuid.Nil || value.ActorID == uuid.Nil ||
			value.CausationID == uuid.Nil || !databaseJobEventType.MatchString(value.EventType) ||
			!databaseJobEventType.MatchString(value.AggregateType) || value.EventVersion < 1 ||
			value.AggregateVersion < 1 || value.EventSequence < 1 ||
			value.OccurredAt.IsZero() || value.OccurredAt.Location() != time.UTC ||
			value.OccurredAt.Nanosecond()%int(time.Microsecond) != 0 ||
			!databaseJobCorrelation.MatchString(value.CorrelationID) ||
			!validDatabaseJobText(value.IdempotencyKey, eventsv1.MaximumKeyBytes) ||
			!validDatabaseJobText(
				value.BusinessEffectIdempotencyKey, eventsv1.MaximumKeyBytes,
			) ||
			!databaseJobDigest.MatchString(value.PayloadSchemaSHA256) ||
			hex.EncodeToString(digest[:]) != value.PayloadSHA256 || seen[value.EventID] {
			return false
		}
		seen[value.EventID] = true
	}
	return true
}

func (value DatabaseJobEventInput) clone() DatabaseJobEventInput {
	value.Payload = append(json.RawMessage(nil), value.Payload...)
	return value
}
