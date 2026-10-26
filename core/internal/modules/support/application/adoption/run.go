package adoption

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

const transformVersion = "support-v1-to-v2/v1"

func TransformSHA256() [32]byte { return sha256.Sum256([]byte(transformVersion)) }

func (run Run) valid() bool {
	if run.OperationID == uuid.Nil || run.RunID == uuid.Nil || run.RunID.Version() != 7 ||
		run.SourceRelationOID == 0 || run.ExpectedCount < 0 || run.CopiedCount < 0 ||
		run.CopiedCount > run.ExpectedCount || run.Revision < 1 || run.NextPage < 1 ||
		run.LeaseGeneration < 0 || run.CreatedAt.IsZero() || run.UpdatedAt.Before(run.CreatedAt) ||
		zeroSHA(run.LocatorSHA256) || zeroSHA(run.ManifestSHA256) ||
		zeroSHA(run.TransformSHA256) || zeroSHA(run.ExpectedSourceSHA256) ||
		zeroSHA(run.SourceSHA256) || zeroSHA(run.TargetSHA256) {
		return false
	}
	if run.CopiedCount > 0 && run.CursorID == uuid.Nil {
		return false
	}
	switch run.Status {
	case RunPlanned, RunApplying, RunReconciling, RunCompleted, RunFenced:
	default:
		return false
	}
	_, err := run.locator()
	return err == nil
}

func (run Run) locator() (VerifiedLocator, error) {
	return NewVerifiedLocator(run.Kind, run.TenantID, run.SourceSchema,
		run.SourceRelation, run.SourceRelationOID, run.TenantSlug, run.TenantName,
		run.LocatorSHA256)
}

func exactManifest(left, right SourceManifest) bool {
	return left.SnapshotID == right.SnapshotID && left.RowCount == right.RowCount &&
		left.RootSHA256 == right.RootSHA256 && left.ManifestSHA256 == right.ManifestSHA256
}

func validObservedAt(value time.Time) bool {
	return !value.IsZero() && value.Nanosecond()%1000 == 0
}
