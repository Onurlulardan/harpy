package platformlicense

import (
	"time"

	"github.com/google/uuid"
)

type Limit struct {
	Unlimited bool  `json:"unlimited"`
	Value     int64 `json:"value"`
}

type License struct {
	LicenseID        uuid.UUID
	EnvironmentID    uuid.UUID
	SnapshotSequence int64
	SnapshotSHA256   string
	State            string
	ReleaseChannel   string
	Features         map[string]bool
	Limits           map[string]Limit
	DatabaseNow      time.Time
}

type RefreshCommand struct {
	EnvironmentID  uuid.UUID
	TenantID       uuid.UUID
	ActorID        uuid.UUID
	IdempotencyKey string
	RequestSHA256  string
}

type RefreshAccepted struct {
	OperationID uuid.UUID
	Replayed    bool
	CreatedAt   time.Time
}

type RefreshOperation struct {
	OperationID      uuid.UUID
	State            string
	SnapshotSequence *int64
	SnapshotSHA256   *string
	FailureCode      *string
	CreatedAt        time.Time
	CompletedAt      *time.Time
}
