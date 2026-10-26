package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type publishClaimRow struct {
	eventID          uuid.UUID
	tenantID         pgtype.UUID
	scopeKind        string
	scopeID          uuid.UUID
	groupID          string
	groupGeneration  int64
	domain           string
	generation       int64
	writerOrigin     string
	contractVersion  string
	eventType        string
	eventVersion     int32
	aggregateType    string
	aggregateID      uuid.UUID
	aggregateVersion int64
	eventSequence    int64
	occurredAt       time.Time
	actorID          uuid.UUID
	correlationID    string
	causationID      uuid.UUID
	idempotencyKey   string
	businessKey      string
	payload          []byte
	payloadSHA256    string
	schemaSHA256     string
	deliveryAttempts int32
	leaseOwnerID     uuid.UUID
	leaseGeneration  int64
	heartbeatAt      time.Time
	leaseUntil       time.Time
	nextAttemptAt    time.Time
}
