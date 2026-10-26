package adoption

import (
	"regexp"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type SourceKind string
type Status string

const (
	ZeroSource  SourceKind = "zero"
	LegacyTrace SourceKind = "legacy_trace_logs"

	StatusPlanned     Status = "planned"
	StatusApplying    Status = "applying"
	StatusReconciling Status = "reconciling"
	StatusCompleted   Status = "completed"
	StatusFenced      Status = "fenced"
)

var relationPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type Manifest struct {
	Kind             SourceKind
	Schema, Relation string
	RelationOID      uint32
	RowCount         int64
	RootSHA256       [32]byte
	SnapshotID       string
}

func (value Manifest) Validate() error {
	zero := value.Kind == ZeroSource && value.RelationOID == 0 && value.RowCount == 0
	legacy := value.Kind == LegacyTrace && value.RelationOID > 0 && value.RowCount >= 0
	if (!zero && !legacy) || !relationPattern.MatchString(value.Schema) ||
		!relationPattern.MatchString(value.Relation) || value.RootSHA256 == [32]byte{} ||
		len(value.SnapshotID) < 16 || len(value.SnapshotID) > 200 {
		return domain.ErrInvalid
	}
	return nil
}

func (value Manifest) Same(other Manifest) bool {
	return value == other
}

type Run struct {
	RunID, OperationID          uuid.UUID
	Manifest                    Manifest
	TransformSHA256             [32]byte
	NextPage, CopiedCount       int64
	CursorID                    *uuid.UUID
	Status                      Status
	LeaseOwner                  *uuid.UUID
	LeaseGeneration, Revision   int64
	LeaseExpiresAt, CompletedAt *time.Time
}

func (value Run) Validate() error {
	if value.RunID == uuid.Nil || value.RunID.Version() != 7 || value.OperationID == uuid.Nil ||
		value.Manifest.Validate() != nil || value.TransformSHA256 == [32]byte{} ||
		value.NextPage < 1 || value.CopiedCount < 0 ||
		value.CopiedCount > value.Manifest.RowCount || value.Revision < 1 ||
		value.LeaseGeneration < 0 || !validStatus(value.Status) ||
		(value.LeaseOwner == nil) != (value.LeaseExpiresAt == nil) ||
		(value.Status == StatusCompleted) != (value.CompletedAt != nil) {
		return domain.ErrUnavailable
	}
	return nil
}

func validStatus(value Status) bool {
	return value == StatusPlanned || value == StatusApplying || value == StatusReconciling ||
		value == StatusCompleted || value == StatusFenced
}
