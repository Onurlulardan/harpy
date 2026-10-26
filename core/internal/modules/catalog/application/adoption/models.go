package adoption

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	SourceCount, TargetCount           int64
	SourceRootSHA256, TargetRootSHA256 string
	Blockers                           []string
	RegistryReferences                 []RegistryReference
}

type RegistryReference struct {
	Kind, Identity           string
	PackageID, Slug, Version string
	ManifestSHA256           string
}

type Run struct {
	ID                                    uuid.UUID
	SourceManifestSHA256, TransformSHA256 string
	SourceCount, TargetCount              int64
	SourceRootSHA256, TargetRootSHA256    string
	State, BlockerCode                    string
	CutoverGeneration                     int64
	CreatedAt                             time.Time
	VerifiedAt                            *time.Time
}

type VerifyRequest struct {
	RunID                                 uuid.UUID
	SourceManifestSHA256, TransformSHA256 string
	CutoverGeneration                     int64
}

type RouteRequest struct {
	RouteKey          string
	RequiredConsumers int64
	RemoveAfter       time.Time
}

type ConsumerAck struct {
	RouteKey, ConsumerID, EvidenceSHA256 string
}
