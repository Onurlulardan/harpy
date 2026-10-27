package adoption

import (
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/google/uuid"
)

type runDocument struct {
	RunID             uuid.UUID      `json:"run_id"`
	OperationID       uuid.UUID      `json:"operation_id"`
	SourceKind        app.SourceKind `json:"source_kind"`
	SourceSchema      string         `json:"source_schema"`
	SourceRelation    string         `json:"source_relation"`
	SourceRelationOID uint32         `json:"source_relation_oid"`
	SourceCount       int64          `json:"source_count"`
	SourceRootSHA256  string         `json:"source_root_sha256"`
	SnapshotID        string         `json:"snapshot_id"`
	TransformSHA256   string         `json:"transform_sha256"`
	NextPage          int64          `json:"next_page"`
	CursorID          *uuid.UUID     `json:"cursor_id"`
	CopiedCount       int64          `json:"copied_count"`
	Status            app.Status     `json:"status"`
	LeaseOwner        *uuid.UUID     `json:"lease_owner"`
	LeaseGeneration   int64          `json:"lease_generation"`
	LeaseExpiresAt    *time.Time     `json:"lease_expires_at"`
	Revision          int64          `json:"revision"`
	CompletedAt       *time.Time     `json:"completed_at"`
}

type receiptDocument struct {
	RunID            uuid.UUID  `json:"run_id"`
	PageNumber       int64      `json:"page_number"`
	CursorBefore     *uuid.UUID `json:"cursor_before"`
	CursorAfter      uuid.UUID  `json:"cursor_after"`
	RowCount         int        `json:"row_count"`
	SourcePageSHA256 string     `json:"source_page_sha256"`
	TargetPageSHA256 string     `json:"target_page_sha256"`
	State            string     `json:"state"`
	AppliedAt        time.Time  `json:"applied_at"`
	AdvancedAt       *time.Time `json:"advanced_at"`
}
