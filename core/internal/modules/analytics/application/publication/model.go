package publication

import (
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	"github.com/google/uuid"
)

type ResourceKind string

const (
	ResourceQuery     ResourceKind = "query"
	ResourceDashboard ResourceKind = "dashboard"
)

func (value ResourceKind) Valid() bool { return value == ResourceQuery || value == ResourceDashboard }

type Snapshot struct {
	ReleaseID               uuid.UUID
	ResourceKind            ResourceKind
	ResourceVersionID       uuid.UUID
	Audience                analyticspublication.Audience
	State                   analyticspublication.State
	CachePolicy             analyticspublication.CachePolicy
	ManifestSHA256          [32]byte
	RequesterID, ApproverID uuid.UUID
	Revision, Generation    uint64
	PublicID                [16]byte
	CapabilityExpiresAt     *time.Time
	CreatedAt, UpdatedAt    time.Time
}

type Page struct {
	Items []Snapshot
	After *uuid.UUID
}

type Activation struct {
	Snapshot        Snapshot
	CapabilityToken []byte
	Replay          bool
}

type PublicRoute struct {
	TenantID, ReleaseID uuid.UUID
	ResourceKind        ResourceKind
	ResourceVersionID   uuid.UUID
	Audience            analyticspublication.Audience
	State               analyticspublication.State
	Generation          uint64
	PublicID            [16]byte
	CapabilitySHA256    [32]byte
	CapabilityExpiresAt *time.Time
}

type PublicRequest struct {
	PublicID        [16]byte
	ExpectedKind    ResourceKind
	Slug            string
	TrustedClientIP string
	CapabilityToken []byte
	Parameters      []byte
}

type PublicResult struct{ Body []byte }
