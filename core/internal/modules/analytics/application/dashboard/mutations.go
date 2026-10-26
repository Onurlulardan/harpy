package dashboard

import (
	"time"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	"github.com/google/uuid"
)

type CreateMutation struct {
	OperationID, ActorID, DashboardID, VersionID uuid.UUID
	IdempotencyKey                               string
	RequestSHA256                                [32]byte
	Name                                         string
	Visibility                                   analyticsdashboard.Visibility
	Layout                                       Document
	Widgets                                      []Widget
	OccurredAt                                   time.Time
}

type UpdateMutation struct {
	OperationID, ActorID, DashboardID             uuid.UUID
	IdempotencyKey                                string
	RequestSHA256                                 [32]byte
	ExpectedRootRevision, ExpectedVersionRevision uint64
	Name                                          string
	Visibility                                    analyticsdashboard.Visibility
	Layout                                        Document
	Widgets                                       []Widget
	OccurredAt                                    time.Time
}

type DeleteMutation struct {
	OperationID, ActorID, DashboardID uuid.UUID
	IdempotencyKey                    string
	RequestSHA256                     [32]byte
	ExpectedRootRevision              uint64
	OccurredAt                        time.Time
}

type PublishMutation struct {
	OperationID, ActorID, DashboardID, DraftVersionID,
	NextDraftVersionID uuid.UUID
	IdempotencyKey                              string
	RequestSHA256                               [32]byte
	ExpectedRootRevision, ExpectedDraftRevision uint64
	ManifestSHA256                              [32]byte
	OccurredAt                                  time.Time
}

type ACLMutation struct {
	OperationID, ActorID, DashboardID, PrincipalID uuid.UUID
	PrincipalKind                                  string
	IdempotencyKey                                 string
	RequestSHA256                                  [32]byte
	Level                                          string
	ExpectedRevision                               uint64
	OccurredAt                                     time.Time
}

type DefaultMutation struct {
	OperationID, ActorID, PrincipalID, DashboardID uuid.UUID
	IdempotencyKey                                 string
	RequestSHA256                                  [32]byte
	Surface                                        string
	OccurredAt                                     time.Time
}
