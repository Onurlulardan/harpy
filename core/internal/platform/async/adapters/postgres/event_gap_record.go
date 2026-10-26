package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
)

type eventGapRecord struct {
	rootID     uuid.UUID
	jobID      uuid.UUID
	canonical  []byte
	sourceSHA  string
	errorSHA   string
	redacted   []byte
	payloadRef string
	dedupeKey  string
	evidence   eventsv1.SequenceGapEvidence
	consumer   eventsv1.ConsumerPolicy
}

func newEventGapRecord(
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	expected int64,
) (eventGapRecord, error) {
	evidence, err := eventsv1.NewSequenceGapEvidence(
		envelope, repository.consumer, repository.consumerVersion, expected,
		repository.producers, repository.consumers,
	)
	if err != nil {
		return eventGapRecord{}, err
	}
	document, err := json.Marshal(evidence)
	if err != nil {
		return eventGapRecord{}, fmt.Errorf("marshal event gap evidence: %w", err)
	}
	canonical, err := canonicaljsonv1.Canonicalize(document, eventsv1.MaximumDocumentBytes)
	if err != nil {
		return eventGapRecord{}, fmt.Errorf("canonicalize event gap evidence: %w", err)
	}
	sourceSHA := digestEventGap(canonical)
	rootID := deterministicEventGapID(
		"event-gap-root.v1", envelope.EventID, repository.consumer,
		fmt.Sprint(repository.consumerVersion), sourceSHA,
	)
	jobID := deterministicEventGapID("event-gap-job.v1", rootID.String())
	redacted, err := marshalEventGapMetadata(evidence)
	if err != nil {
		return eventGapRecord{}, err
	}
	consumer, _ := repository.consumers.Resolve(
		repository.consumer, repository.consumerVersion,
		envelope.EventType, envelope.EventVersion,
	)
	return eventGapRecord{
		rootID: rootID, jobID: jobID, canonical: canonical,
		sourceSHA: sourceSHA,
		errorSHA: digestEventGap([]byte(fmt.Sprintf(
			"event-sequence-gap.v1\n%s\n%s\n%d\n%d\n%d",
			envelope.EventID, repository.consumer, repository.consumerVersion,
			expected, envelope.EventSequence,
		))),
		redacted: redacted, payloadRef: "deadletter:event-gap:" + rootID.String(),
		dedupeKey: "event-gap:" + repository.consumer + ":" +
			fmt.Sprint(repository.consumerVersion) + ":" + envelope.EventID,
		evidence: evidence, consumer: consumer,
	}, nil
}

func deterministicEventGapID(parts ...string) uuid.UUID {
	digest := sha256.New()
	for _, part := range parts {
		digest.Write([]byte(part))
		digest.Write([]byte{0})
	}
	bytes := digest.Sum(nil)[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x80
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return uuid.Must(uuid.FromBytes(bytes))
}

func digestEventGap(document []byte) string {
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}

func eventGapReason() (deadletter.ReasonClass, string) {
	return deadletter.ReasonSequenceGap, deadletter.EventSequenceCode
}
