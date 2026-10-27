package legalhold

import (
	"time"

	"github.com/google/uuid"
)

type PlaceRequest struct {
	TenantID       uuid.UUID
	ActorID        uuid.UUID
	IdempotencyKey string
	Reason         string
	Ticket         string
}

type ReleaseRequest struct {
	TenantID       uuid.UUID
	HoldID         uuid.UUID
	ActorID        uuid.UUID
	IdempotencyKey string
	Reason         string
}

type PlaceMutation struct {
	HoldID      uuid.UUID
	OperationID uuid.UUID
	TenantID    uuid.UUID
	ActorID     uuid.UUID
	Key         string
	RequestSHA  []byte
	Reason      string
	Ticket      string
	OccurredAt  time.Time
}

type ReleaseMutation struct {
	OperationID uuid.UUID
	TenantID    uuid.UUID
	HoldID      uuid.UUID
	ActorID     uuid.UUID
	Key         string
	RequestSHA  []byte
	Reason      string
	OccurredAt  time.Time
}

type Snapshot struct {
	HoldID             uuid.UUID
	TenantID           uuid.UUID
	State              string
	PlacedBy           uuid.UUID
	PlaceOperationID   uuid.UUID
	Reason             string
	Ticket             string
	PlacedAt           time.Time
	ReleasedBy         uuid.UUID
	ReleaseOperationID uuid.UUID
	ReleaseReason      string
	ReleasedAt         time.Time
	Revision           int64
	Replay             bool
}
