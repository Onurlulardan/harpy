package async

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	AnalyticsDashboardExportJobKind       = "analytics.dashboard.export.render"
	AnalyticsDashboardExportMaxAttempts   = int64(3)
	AnalyticsDashboardExportPayloadRev    = int64(1)
	AnalyticsDashboardExportReserveBytes  = int64(2147483648)
	AnalyticsDashboardExportTenantMaxByte = int64(21474836480)
)

var (
	ErrAnalyticsDashboardExportSourceUnavailable = errors.New(
		"analytics dashboard export source is unavailable",
	)
	ErrAnalyticsDashboardExportEvidenceInvalid = errors.New(
		"analytics dashboard export evidence is invalid",
	)
	analyticsDashboardExportSchemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

type AnalyticsDashboardExportCandidate struct {
	TenantID          uuid.UUID
	TenantSchema      string
	ExportID          uuid.UUID
	JobID             uuid.UUID
	AttemptGeneration int64
}

func (candidate AnalyticsDashboardExportCandidate) Valid() bool {
	return candidate.TenantID != uuid.Nil && validAnalyticsExportSchema(
		candidate.TenantSchema,
	) && validAnalyticsExportV7(candidate.ExportID) &&
		validAnalyticsExportV7(candidate.JobID) &&
		candidate.AttemptGeneration >= 1 &&
		candidate.AttemptGeneration <= AnalyticsDashboardExportMaxAttempts
}

type AnalyticsDashboardExportUnknownCandidate struct {
	TenantID            uuid.UUID
	TenantSchema        string
	ExportID            uuid.UUID
	JobID               uuid.UUID
	AttemptID           uuid.UUID
	AttemptGeneration   int64
	LeaseOwnerID        uuid.UUID
	EffectID            uuid.UUID
	EffectRequestSHA256 [32]byte
	UnknownAt           time.Time
}

func (candidate AnalyticsDashboardExportUnknownCandidate) Valid() bool {
	return candidate.TenantID != uuid.Nil && validAnalyticsExportSchema(
		candidate.TenantSchema,
	) && validAnalyticsExportV7(candidate.ExportID) &&
		validAnalyticsExportV7(candidate.JobID) &&
		validAnalyticsExportV7(candidate.AttemptID) &&
		candidate.AttemptGeneration >= 1 &&
		candidate.AttemptGeneration <= AnalyticsDashboardExportMaxAttempts &&
		validAnalyticsExportV7(candidate.LeaseOwnerID) &&
		validAnalyticsExportV7(candidate.EffectID) &&
		candidate.EffectRequestSHA256 != [32]byte{} && !candidate.UnknownAt.IsZero()
}

type AnalyticsDashboardExportExpiredCandidate struct {
	TenantID            uuid.UUID
	TenantSchema        string
	ExportID            uuid.UUID
	JobID               uuid.UUID
	AttemptID           uuid.UUID
	AttemptGeneration   int64
	LeaseOwnerID        uuid.UUID
	EffectStarted       bool
	EffectID            uuid.UUID
	EffectRequestSHA256 [32]byte
	CancellationPending bool
	LeaseExpiredAt      time.Time
}

func (candidate AnalyticsDashboardExportExpiredCandidate) Valid() bool {
	baseValid := candidate.TenantID != uuid.Nil && validAnalyticsExportSchema(
		candidate.TenantSchema,
	) && validAnalyticsExportV7(candidate.ExportID) &&
		validAnalyticsExportV7(candidate.JobID) &&
		validAnalyticsExportV7(candidate.AttemptID) &&
		candidate.AttemptGeneration >= 1 &&
		candidate.AttemptGeneration <= AnalyticsDashboardExportMaxAttempts &&
		validAnalyticsExportV7(candidate.LeaseOwnerID) &&
		!candidate.LeaseExpiredAt.IsZero()
	if !baseValid {
		return false
	}
	if !candidate.EffectStarted {
		return candidate.EffectID == uuid.Nil &&
			candidate.EffectRequestSHA256 == [32]byte{}
	}
	return validAnalyticsExportV7(candidate.EffectID) &&
		candidate.EffectRequestSHA256 != [32]byte{}
}

type AnalyticsDashboardExportSpoolState string

const (
	AnalyticsDashboardExportSpoolInactive AnalyticsDashboardExportSpoolState = "inactive"
	AnalyticsDashboardExportSpoolRunning  AnalyticsDashboardExportSpoolState = "running"
	AnalyticsDashboardExportSpoolUnknown  AnalyticsDashboardExportSpoolState = "outcome_unknown"
)

func (state AnalyticsDashboardExportSpoolState) Valid() bool {
	return state == AnalyticsDashboardExportSpoolInactive ||
		state == AnalyticsDashboardExportSpoolRunning ||
		state == AnalyticsDashboardExportSpoolUnknown
}

type AnalyticsDashboardExportWorkerSource interface {
	NextAnalyticsDashboardExport(
		context.Context,
	) (AnalyticsDashboardExportCandidate, bool, error)
	NextAnalyticsDashboardExportUnknown(
		context.Context,
	) (AnalyticsDashboardExportUnknownCandidate, bool, error)
	NextAnalyticsDashboardExportExpired(
		context.Context,
	) (AnalyticsDashboardExportExpiredCandidate, bool, error)
	AnalyticsDashboardExportSpoolTenant(
		context.Context, uuid.UUID,
	) (uuid.UUID, bool, error)
}

func validAnalyticsExportV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7) &&
		value.Variant() == uuid.RFC4122
}

func validAnalyticsExportSchema(value string) bool {
	return len(value) >= 1 && len(value) <= 63 &&
		analyticsDashboardExportSchemaPattern.MatchString(value)
}
