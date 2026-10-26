package postgres

import (
	"time"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
)

type deadLetterRootRow struct {
	id                uuid.UUID
	tenantID          uuid.UUID
	scopeKind         string
	scopeID           uuid.UUID
	groupID           string
	groupGeneration   int64
	domain            string
	generation        int64
	sourceKind        deadletter.SourceKind
	sourceID          uuid.UUID
	consumer          *string
	consumerVersion   *int
	sourceContract    string
	sourceVersion     int
	sourceSHA256      string
	payload           []byte
	payloadSHA256     string
	payloadRedactedAt *time.Time
	redactedEnvelope  []byte
	payloadExpiresAt  time.Time
	metadataExpiresAt time.Time
	observedAt        time.Time
	nextSequence      int
}

func (row deadLetterRootRow) retryable(tenantID uuid.UUID) error {
	if row.id == uuid.Nil || row.tenantID != tenantID || row.scopeKind != "tenant" ||
		row.scopeID != tenantID || row.sourceID == uuid.Nil {
		return ErrDeadLetterNotFound
	}
	if row.payload == nil || row.payloadRedactedAt != nil ||
		!row.observedAt.Before(row.payloadExpiresAt) {
		return ErrDeadLetterPayloadExpired
	}
	if row.nextSequence < 1 || !row.observedAt.Before(row.metadataExpiresAt) {
		return ErrDeadLetterRetryInvalid
	}
	return nil
}
