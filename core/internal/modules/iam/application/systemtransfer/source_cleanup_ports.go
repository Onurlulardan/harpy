package systemtransfer

import (
	"context"
	"fmt"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

const (
	SourceCleanupOutcomeUnknown = "SOURCE_DELETE_OUTCOME_UNKNOWN"
	SourceCleanupUnavailable    = "SOURCE_DELETE_UNAVAILABLE"
)

type SourceCleanupClaim struct {
	AfterTenantID uuid.UUID
	LeaseOwnerID  uuid.UUID
	ObservedAt    time.Time
}

type SourceCleanupTarget struct {
	TenantID, ActorID, SourceID uuid.UUID
	Resource                    publicresource.Resource
	ObjectKey, ObjectVersion    string
	LeaseOwnerID                uuid.UUID
	Attempt                     int
	LeaseGeneration             int64
}

func (target SourceCleanupTarget) Validate() error {
	if target.TenantID == uuid.Nil || target.ActorID == uuid.Nil ||
		target.SourceID.Version() != uuid.Version(7) || !target.Resource.Valid() ||
		target.ObjectKey != importSourceObjectKey(target.SourceID) ||
		len(target.ObjectVersion) < 1 || len(target.ObjectVersion) > 512 ||
		target.LeaseOwnerID.Version() != uuid.Version(7) || target.Attempt < 1 ||
		target.LeaseGeneration < 1 {
		return ErrSourceCleanupInvalid
	}
	return nil
}

func importSourceObjectKey(sourceID uuid.UUID) string {
	return fmt.Sprintf("iam-system-import-sources/v1/%s/1.fcart", sourceID)
}

type SourceCleanupScan struct {
	ScannedTenantID uuid.UUID
	Target          SourceCleanupTarget
	Found           bool
}

func (scan SourceCleanupScan) Validate() error {
	if scan.ScannedTenantID == uuid.Nil {
		if scan.Found || scan.Target != (SourceCleanupTarget{}) {
			return ErrSourceCleanupInvalid
		}
		return nil
	}
	if scan.Found && (scan.Target.Validate() != nil ||
		scan.Target.TenantID != scan.ScannedTenantID) ||
		!scan.Found && scan.Target != (SourceCleanupTarget{}) {
		return ErrSourceCleanupInvalid
	}
	return nil
}

type SourceDeleteResult struct {
	ProviderRequestID    string
	DeleteResponseSHA256 [32]byte
	AbsenceProbeSHA256   [32]byte
	VerifiedAbsentAt     time.Time
}

func (result SourceDeleteResult) Validate() error {
	if len(result.ProviderRequestID) < 1 || len(result.ProviderRequestID) > 256 ||
		result.DeleteResponseSHA256 == [32]byte{} ||
		result.AbsenceProbeSHA256 == [32]byte{} ||
		!controlTime(result.VerifiedAbsentAt) {
		return ErrSourceCleanupInvalid
	}
	return nil
}

type SourceCleanupCompletion struct {
	Target  SourceCleanupTarget
	Receipt SourceDeleteResult
}

type SourceCleanupRetry struct {
	Target     SourceCleanupTarget
	ErrorCode  string
	ObservedAt time.Time
	RetryAt    time.Time
}

type SourceCleanupStore interface {
	AcquireSourceCleanup(context.Context, SourceCleanupClaim) (SourceCleanupScan, error)
	PersistSourceCleanupCompletion(context.Context, SourceCleanupCompletion) (bool, error)
	ScheduleSourceCleanupRetry(context.Context, SourceCleanupRetry) (bool, error)
}

type SourceObjectDeleter interface {
	DeleteExactImportSource(context.Context, string, string) (
		SourceDeleteResult, error)
}

type SourceCleanupClock interface{ Now() time.Time }
type SourceCleanupIDs interface {
	NewV7(time.Time) (uuid.UUID, error)
}
type SourceCleanupJitter interface{ Uint64N(uint64) uint64 }
