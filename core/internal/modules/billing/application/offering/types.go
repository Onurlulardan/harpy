package offering

import (
	"time"

	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
	"github.com/google/uuid"
)

type CreateRequest struct {
	ActorID        uuid.UUID
	IdempotencyKey string
	Key            offeringdomain.Key
	Version        offeringdomain.Version
}

type CreateVersionRequest struct {
	ActorID          uuid.UUID
	IdempotencyKey   string
	OfferingID       uuid.UUID
	ExpectedRevision int64
	Version          offeringdomain.Version
}

type RetireRequest struct {
	ActorID          uuid.UUID
	IdempotencyKey   string
	OfferingID       uuid.UUID
	ExpectedRevision int64
	Reason           string
}

type CreateMutation struct {
	OperationID    uuid.UUID
	OfferingID     uuid.UUID
	VersionID      uuid.UUID
	PriceIDs       []uuid.UUID
	ActorID        uuid.UUID
	IdempotencyKey string
	RequestSHA256  [32]byte
	Key            offeringdomain.Key
	Version        offeringdomain.Version
	OccurredAt     time.Time
}

type CreateVersionMutation struct {
	OperationID      uuid.UUID
	OfferingID       uuid.UUID
	VersionID        uuid.UUID
	PriceIDs         []uuid.UUID
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    [32]byte
	ExpectedRevision int64
	Version          offeringdomain.Version
	OccurredAt       time.Time
}

type RetireMutation struct {
	OperationID      uuid.UUID
	OfferingID       uuid.UUID
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    [32]byte
	ExpectedRevision int64
	Reason           string
	OccurredAt       time.Time
}

type Snapshot struct {
	OfferingID uuid.UUID
	VersionID  uuid.UUID
	Revision   int64
	Version    int64
	Lifecycle  string
	Replay     bool
}
