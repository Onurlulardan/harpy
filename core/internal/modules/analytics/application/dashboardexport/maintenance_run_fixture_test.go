package dashboardexport

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type maintenanceRecoveryStoreFixture struct {
	candidate                                    MaintenanceUnknownCandidate
	claim                                        MaintenanceUnknownClaim
	found                                        bool
	deferred                                     bool
	resolved                                     *MaintenanceResolution
	cancel                                       context.CancelFunc
	expiredScans, unknownScans, maintenanceScans int
}

func (*maintenanceRecoveryStoreFixture) ActiveSpoolReservation(
	context.Context, SpoolReservationIdentity,
) (bool, error) {
	return false, nil
}

func (store *maintenanceRecoveryStoreFixture) NextUnknown(
	context.Context,
) (UnknownCandidate, bool, error) {
	store.unknownScans++
	return UnknownCandidate{}, false, nil
}

func (*maintenanceRecoveryStoreFixture) ClaimUnknown(
	context.Context, UnknownCandidate, uuid.UUID, time.Time,
) (UnknownCleanupClaim, error) {
	return UnknownCleanupClaim{}, ErrUnavailable
}

func (*maintenanceRecoveryStoreFixture) CheckpointUnknown(
	context.Context, UnknownCleanupClaim, time.Time,
) (UnknownCleanupLease, error) {
	return UnknownCleanupLease{}, ErrUnavailable
}

func (*maintenanceRecoveryStoreFixture) DeferUnknown(
	context.Context, UnknownCleanupClaim, time.Time,
) (time.Time, error) {
	return time.Time{}, ErrUnavailable
}

func (*maintenanceRecoveryStoreFixture) ResolveUnknown(
	context.Context, UnknownCleanupClaim, UnknownResolution,
) (State, error) {
	return "", ErrUnavailable
}

func (store *maintenanceRecoveryStoreFixture) NextExpired(
	context.Context,
) (ExpiredCandidate, bool, error) {
	store.expiredScans++
	return ExpiredCandidate{}, false, nil
}

func (*maintenanceRecoveryStoreFixture) FinalizeExpired(
	context.Context, ExpiredCandidate, time.Time,
) (ExpiredFinalization, error) {
	return ExpiredFinalization{}, ErrUnavailable
}

func (store *maintenanceRecoveryStoreFixture) NextMaintenanceUnknown(
	context.Context,
) (MaintenanceUnknownCandidate, bool, error) {
	store.maintenanceScans++
	return store.candidate, store.found, nil
}

func (store *maintenanceRecoveryStoreFixture) ClaimMaintenanceUnknown(
	_ context.Context, candidate MaintenanceUnknownCandidate,
	cleanupOwner uuid.UUID, _ time.Time,
) (MaintenanceUnknownClaim, error) {
	claim := store.claim
	claim.Candidate = candidate
	claim.CleanupLeaseOwnerID = cleanupOwner
	return claim, nil
}

func (*maintenanceRecoveryStoreFixture) CheckpointMaintenanceUnknown(
	context.Context, MaintenanceUnknownClaim, time.Time,
) (UnknownCleanupLease, error) {
	return UnknownCleanupLease{}, ErrUnavailable
}

func (store *maintenanceRecoveryStoreFixture) DeferMaintenanceUnknown(
	context.Context, MaintenanceUnknownClaim, time.Time,
) (time.Time, error) {
	store.deferred = true
	return time.Now().UTC(), nil
}

func (store *maintenanceRecoveryStoreFixture) ResolveMaintenanceUnknown(
	_ context.Context, _ MaintenanceUnknownClaim, resolution MaintenanceResolution,
) (State, error) {
	store.resolved = &resolution
	if store.cancel != nil {
		store.cancel()
	}
	return resolution.TerminalState, nil
}

func maintenanceCandidateFixture(kind string, base time.Time) MaintenanceUnknownCandidate {
	return MaintenanceUnknownCandidate{
		Kind:             kind,
		TenantID:         uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		TenantSchema:     "tenant_analytics",
		ExportID:         uuid.MustParse("019b1000-0000-7000-8000-000000000021"),
		AuthorityGroupID: "analytics", AuthorityGroupGeneration: 1,
		CutoverGeneration: 1,
		JobID:             uuid.MustParse("019b1000-0000-7000-8000-000000000022"),
		AttemptID:         uuid.MustParse("019b1000-0000-7000-8000-000000000023"),
		AttemptNumber:     1, LeaseGeneration: 1,
		OriginalLeaseOwnerID: uuid.MustParse(
			"019b1000-0000-7000-8000-000000000024"),
		EffectID:            uuid.NewSHA1(uuid.NameSpaceOID, []byte(kind)),
		EffectRequestSHA256: [32]byte{1}, UnknownAt: base,
	}
}
