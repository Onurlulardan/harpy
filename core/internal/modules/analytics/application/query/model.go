package query

import (
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	"github.com/google/uuid"
)

type Snapshot struct {
	QueryID, VersionID uuid.UUID
	VersionNumber      uint64
	RootRevision       uint64
	ReferenceEpoch     uint64
	VersionRevision    uint64
	Lifecycle          analyticsquery.Lifecycle
	VersionState       analyticsquery.VersionState
	CachePolicy        analyticspublication.CachePolicy
	Name               string
	QueryDocument      []byte
	ParameterDocument  []byte
	OutputDocument     []byte
	ManifestSHA256     [32]byte
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Page struct {
	Items []Snapshot
	After *uuid.UUID
}

type DependencyRequest struct {
	OwnerDomain, ResourceKind string
	ResourceID, VersionID     uuid.UUID
}

type DependencySnapshot struct {
	DependencyRequest
	CatalogVersion, QueryContractVersion, PolicyVersion uint64
	OwnerGeneration                                     uint64
	WriterV2                                            bool
	Retired                                             bool
	DataRevisionCapability                              string
	DependencySHA256                                    [32]byte
}

type NormalizedContract struct {
	Query, Parameters, Output Document
}

type ExecutionBinding struct {
	ReleaseGeneration   uint64
	AuthorizationSHA256 [32]byte
	PredicateSHA256     [32]byte
	RevisionsSHA256     [32]byte
}

type ExecutionResult struct {
	Body []byte
}
