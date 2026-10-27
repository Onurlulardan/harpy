package compatibility

import (
	"time"

	"github.com/google/uuid"
)

type Disposition string

const (
	DispositionAdopted Disposition = "adopted"
	DispositionBlocked Disposition = "blocked"
	DispositionRetired Disposition = "retired"
)

func (value Disposition) Valid() bool {
	return value == DispositionAdopted || value == DispositionBlocked ||
		value == DispositionRetired
}

type AdoptionItem struct {
	RunID, LegacyEndpointID                uuid.UUID
	SourceSHA256, TargetSHA256             [32]byte
	Disposition                            Disposition
	ReasonCode                             string
	QueryID, QueryVersionID, PublicationID uuid.UUID
	CreatedAt                              time.Time
}

type Acceptance struct {
	RunID, TenantID           uuid.UUID
	CutoverGeneration         uint64
	SourceCount, AdoptedCount uint64
	RetiredCount              uint64
	RootSHA256                [32]byte
	AcceptedAt                time.Time
}

type AdoptionAuthority struct {
	AuthorityID, TenantID uuid.UUID
	SchemaName            string
}

type LegacyKey struct{ TenantIdentity, Method, Path, Slug string }

type LegacyMapping struct {
	Key                         LegacyKey
	TenantID, LegacyEndpointID  uuid.UUID
	ReleaseID                   uuid.UUID
	PublicID                    [16]byte
	State                       string
	HitCount                    uint64
	LastHitAt, ConsumersAckedAt *time.Time
	RemoveAfter                 time.Time
}

type RetirementEvidence struct {
	AllTenantsV2 bool
	SHA256       [32]byte
	ObservedAt   time.Time
}
