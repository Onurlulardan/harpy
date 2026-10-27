package dashboardexport

import (
	"time"

	"github.com/google/uuid"
)

type AuthEvidence struct {
	TokenVersion, AuthorizationVersion, AuthorizationClock int64
}

type SubmitRequest struct {
	DashboardID    uuid.UUID
	Format         Format
	Parameters     []byte
	IdempotencyKey string
	Auth           AuthEvidence
}

type SubmitMutation struct {
	OperationID, ExportID, JobID, ExpiryJobID, ActorID, DashboardID uuid.UUID
	Format                                                          Format
	Parameters                                                      []byte
	IdempotencyKey                                                  string
	RequestSHA256                                                   [32]byte
	Auth                                                            AuthEvidence
	MaximumRows, MaximumPlaintextBytes                              int64
	OccurredAt, ExpiresAt                                           time.Time
}

type DeleteRequest struct {
	DashboardID, ExportID uuid.UUID
	IdempotencyKey        string
}

type DeleteMutation struct {
	OperationID, DeleteJobID, ActorID uuid.UUID
	DeleteRequest
	RequestSHA256 [32]byte
	OccurredAt    time.Time
}
