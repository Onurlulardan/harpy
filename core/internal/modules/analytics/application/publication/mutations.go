package publication

import (
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	"github.com/google/uuid"
)

type CreateMutation struct {
	OperationID, ActorID, ReleaseID, ResourceVersionID uuid.UUID
	IdempotencyKey                                     string
	RequestSHA256                                      [32]byte
	ResourceKind                                       ResourceKind
	Audience                                           analyticspublication.Audience
	CachePolicy                                        analyticspublication.CachePolicy
	ManifestSHA256                                     [32]byte
	OccurredAt                                         time.Time
}
type TransitionMutation struct {
	OperationID, ActorID, ReleaseID uuid.UUID
	IdempotencyKey                  string
	RequestSHA256                   [32]byte
	ExpectedRevision                uint64
	StepUpAt                        time.Time
	OccurredAt                      time.Time
}
type ActivateMutation struct {
	TransitionMutation
	PublicID            [16]byte
	CapabilitySHA256    [32]byte
	CapabilityExpiresAt *time.Time
}
