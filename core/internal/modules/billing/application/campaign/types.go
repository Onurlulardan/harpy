package campaign

import (
	"time"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
	"github.com/google/uuid"
)

type CreateRequest struct {
	ActorID        uuid.UUID
	IdempotencyKey string
	Key            campaigndomain.Key
	Publication    campaigndomain.Publication
	Available      bool
}

type CreateVersionRequest struct {
	ActorID          uuid.UUID
	IdempotencyKey   string
	CampaignID       uuid.UUID
	ExpectedRevision int64
	Publication      campaigndomain.Publication
	Available        bool
}

type RetireRequest struct {
	ActorID          uuid.UUID
	IdempotencyKey   string
	CampaignID       uuid.UUID
	ExpectedRevision int64
	Reason           string
}

type CreateMutation struct {
	OperationID    uuid.UUID
	CampaignID     uuid.UUID
	VersionID      uuid.UUID
	ActorID        uuid.UUID
	IdempotencyKey string
	RequestSHA256  [32]byte
	Key            campaigndomain.Key
	Publication    campaigndomain.Publication
	Available      bool
	OccurredAt     time.Time
}

type CreateVersionMutation struct {
	OperationID      uuid.UUID
	CampaignID       uuid.UUID
	VersionID        uuid.UUID
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    [32]byte
	ExpectedRevision int64
	Publication      campaigndomain.Publication
	Available        bool
	OccurredAt       time.Time
}

type RetireMutation struct {
	OperationID      uuid.UUID
	CampaignID       uuid.UUID
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    [32]byte
	ExpectedRevision int64
	Reason           string
	OccurredAt       time.Time
}

type Snapshot struct {
	CampaignID uuid.UUID
	VersionID  uuid.UUID
	Revision   int64
	Version    int64
	Lifecycle  string
	Replay     bool
}
