package discountcatalog

import (
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/google/uuid"
)

type CreateRequest struct {
	ActorID        uuid.UUID
	IdempotencyKey string
	Policy         discountdomain.Policy
	Presentation   discountdomain.Presentation
	Available      bool
	Campaign       CampaignUpdate
}

type CreateVersionRequest struct {
	ActorID          uuid.UUID
	IdempotencyKey   string
	DiscountID       uuid.UUID
	ExpectedRevision int64
	Policy           discountdomain.Policy
	Presentation     discountdomain.Presentation
	Available        *bool
	Campaign         CampaignUpdate
}

type CampaignUpdate struct {
	Replace    bool
	CampaignID *uuid.UUID
}

type RetireRequest struct {
	ActorID          uuid.UUID
	IdempotencyKey   string
	DiscountID       uuid.UUID
	ExpectedRevision int64
	Reason           string
}

type CreateMutation struct {
	OperationID    uuid.UUID
	DiscountID     uuid.UUID
	VersionID      uuid.UUID
	ActorID        uuid.UUID
	IdempotencyKey string
	RequestSHA256  [32]byte
	Policy         discountdomain.Policy
	Presentation   discountdomain.Presentation
	Available      bool
	Campaign       CampaignUpdate
	AssociationIDs [4]uuid.UUID
	OccurredAt     time.Time
}

type CreateVersionMutation struct {
	OperationID      uuid.UUID
	DiscountID       uuid.UUID
	VersionID        uuid.UUID
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    [32]byte
	ExpectedRevision int64
	Policy           discountdomain.Policy
	Presentation     discountdomain.Presentation
	Available        *bool
	Campaign         CampaignUpdate
	AssociationIDs   [4]uuid.UUID
	OccurredAt       time.Time
}

type RetireMutation struct {
	OperationID      uuid.UUID
	DiscountID       uuid.UUID
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    [32]byte
	ExpectedRevision int64
	Reason           string
	OccurredAt       time.Time
}

type Snapshot struct {
	DiscountID        uuid.UUID
	VersionID         uuid.UUID
	PreviousVersionID uuid.UUID
	Revision          int64
	Version           int64
	Lifecycle         string
	Replay            bool
}
