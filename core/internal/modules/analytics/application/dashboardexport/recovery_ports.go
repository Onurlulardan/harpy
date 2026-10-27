package dashboardexport

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RecoveryStore interface {
	SpoolLeaseLookup
	NextUnknown(context.Context) (UnknownCandidate, bool, error)
	ClaimUnknown(
		context.Context, UnknownCandidate, uuid.UUID, time.Time,
	) (UnknownCleanupClaim, error)
	CheckpointUnknown(
		context.Context, UnknownCleanupClaim, time.Time,
	) (UnknownCleanupLease, error)
	DeferUnknown(context.Context, UnknownCleanupClaim, time.Time) (time.Time, error)
	ResolveUnknown(
		context.Context, UnknownCleanupClaim, UnknownResolution,
	) (State, error)
	NextExpired(context.Context) (ExpiredCandidate, bool, error)
	FinalizeExpired(
		context.Context, ExpiredCandidate, time.Time,
	) (ExpiredFinalization, error)
	NextMaintenanceUnknown(context.Context) (MaintenanceUnknownCandidate, bool, error)
	ClaimMaintenanceUnknown(
		context.Context, MaintenanceUnknownCandidate, uuid.UUID, time.Time,
	) (MaintenanceUnknownClaim, error)
	CheckpointMaintenanceUnknown(
		context.Context, MaintenanceUnknownClaim, time.Time,
	) (UnknownCleanupLease, error)
	DeferMaintenanceUnknown(
		context.Context, MaintenanceUnknownClaim, time.Time,
	) (time.Time, error)
	ResolveMaintenanceUnknown(
		context.Context, MaintenanceUnknownClaim, MaintenanceResolution,
	) (State, error)
}
