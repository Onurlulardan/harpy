package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type jobClaimRow struct {
	attemptID        uuid.UUID
	jobID            uuid.UUID
	tenantID         pgtype.UUID
	scopeKind        string
	scopeID          uuid.UUID
	groupID          string
	groupGeneration  int64
	domain           string
	generation       int64
	kind             string
	payloadRevision  int32
	payloadReference pgtype.Text
	payloadSHA256    string
	effectClass      string
	effectKey        pgtype.Text
	priority         int32
	runAt            time.Time
	attempt          int32
	maximumAttempts  int32
	leaseOwnerID     uuid.UUID
	leaseGeneration  int64
	startedAt        time.Time
	leaseUntil       time.Time
}
