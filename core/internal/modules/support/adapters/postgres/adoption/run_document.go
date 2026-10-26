package adoption

import (
	"encoding/json"
	"time"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
	"github.com/google/uuid"
)

type runDocument struct {
	OperationID          uuid.UUID              `json:"operation_id"`
	RunID                uuid.UUID              `json:"run_id"`
	SourceKind           application.SourceKind `json:"source_kind"`
	SourceTenantID       *uuid.UUID             `json:"source_tenant_id"`
	SourceSchema         string                 `json:"source_schema"`
	SourceRelation       string                 `json:"source_relation"`
	SourceRelationOID    uint32                 `json:"source_relation_oid"`
	TenantSlug           string                 `json:"tenant_slug"`
	TenantName           string                 `json:"tenant_name"`
	SnapshotID           string                 `json:"snapshot_id"`
	SourceLocatorSHA256  string                 `json:"source_locator_sha256"`
	SourceManifestSHA256 string                 `json:"source_manifest_sha256"`
	TransformSHA256      string                 `json:"transform_sha256"`
	ExpectedCount        int64                  `json:"expected_count"`
	CopiedCount          int64                  `json:"copied_count"`
	CursorID             *uuid.UUID             `json:"cursor_id"`
	NextPage             int64                  `json:"next_page"`
	ExpectedSourceSHA256 string                 `json:"expected_source_sha256"`
	SourceSHA256         string                 `json:"source_sha256"`
	TargetSHA256         string                 `json:"target_sha256"`
	Status               application.RunStatus  `json:"status"`
	Revision             int64                  `json:"revision"`
	LeaseOwner           *uuid.UUID             `json:"lease_owner"`
	LeaseGeneration      int64                  `json:"lease_generation"`
	LeaseExpiresAt       *time.Time             `json:"lease_expires_at"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

func decodeRun(document []byte) (application.Run, error) {
	var wire runDocument
	if len(document) == 0 || json.Unmarshal(document, &wire) != nil ||
		wire.OperationID == uuid.Nil || wire.RunID == uuid.Nil || wire.RunID.Version() != 7 ||
		wire.SourceRelationOID == 0 || wire.Revision < 1 || wire.NextPage < 1 ||
		wire.ExpectedCount < 0 || wire.CopiedCount < 0 || wire.CopiedCount > wire.ExpectedCount ||
		wire.CreatedAt.IsZero() || wire.UpdatedAt.Before(wire.CreatedAt) {
		return application.Run{}, errDocumentInvalid
	}
	locatorSHA, err := decodeSHA(wire.SourceLocatorSHA256)
	if err != nil {
		return application.Run{}, err
	}
	manifestSHA, err := decodeSHA(wire.SourceManifestSHA256)
	if err != nil {
		return application.Run{}, err
	}
	transformSHA, err := decodeSHA(wire.TransformSHA256)
	if err != nil {
		return application.Run{}, err
	}
	expectedSHA, err := decodeSHA(wire.ExpectedSourceSHA256)
	if err != nil {
		return application.Run{}, err
	}
	sourceSHA, err := decodeSHA(wire.SourceSHA256)
	if err != nil {
		return application.Run{}, err
	}
	targetSHA, err := decodeSHA(wire.TargetSHA256)
	if err != nil {
		return application.Run{}, err
	}
	return application.Run{OperationID: wire.OperationID, RunID: wire.RunID,
		TenantID: valueUUID(wire.SourceTenantID), CursorID: valueUUID(wire.CursorID),
		LeaseOwner: valueUUID(wire.LeaseOwner), Kind: wire.SourceKind,
		SourceSchema: wire.SourceSchema, SourceRelation: wire.SourceRelation,
		SourceRelationOID: wire.SourceRelationOID, TenantSlug: wire.TenantSlug,
		TenantName: wire.TenantName, SnapshotID: wire.SnapshotID,
		LocatorSHA256: locatorSHA, ManifestSHA256: manifestSHA,
		TransformSHA256: transformSHA, ExpectedSourceSHA256: expectedSHA,
		SourceSHA256: sourceSHA, TargetSHA256: targetSHA,
		ExpectedCount: wire.ExpectedCount, CopiedCount: wire.CopiedCount,
		Revision: wire.Revision, NextPage: wire.NextPage,
		LeaseGeneration: wire.LeaseGeneration, Status: wire.Status,
		LeaseExpiresAt: valueTime(wire.LeaseExpiresAt), CreatedAt: wire.CreatedAt,
		UpdatedAt: wire.UpdatedAt}, nil
}
