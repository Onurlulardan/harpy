package query

import (
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	"github.com/google/uuid"
)

type CreateMutation struct {
	OperationID, ActorID, QueryID, VersionID uuid.UUID
	IdempotencyKey                           string
	RequestSHA256                            [32]byte
	Name                                     string
	Query, Parameters, Output                Document
	OccurredAt                               time.Time
}

type UpdateMutation struct {
	OperationID, ActorID, QueryID                 uuid.UUID
	IdempotencyKey                                string
	RequestSHA256                                 [32]byte
	ExpectedRootRevision, ExpectedVersionRevision uint64
	Name                                          string
	Query, Parameters, Output                     Document
	OccurredAt                                    time.Time
}

type DeleteMutation struct {
	OperationID, ActorID, QueryID uuid.UUID
	IdempotencyKey                string
	RequestSHA256                 [32]byte
	ExpectedRootRevision          uint64
	OccurredAt                    time.Time
}

type PublishMutation struct {
	OperationID, ActorID, QueryID, DraftVersionID, NextDraftVersionID uuid.UUID
	IdempotencyKey                                                    string
	RequestSHA256                                                     [32]byte
	ExpectedRootRevision, ExpectedDraftRevision                       uint64
	Contract                                                          NormalizedContract
	Dependencies                                                      []DependencySnapshot
	ManifestSHA256                                                    [32]byte
	CachePolicy                                                       analyticspublication.CachePolicy
	OccurredAt                                                        time.Time
}
