package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type recoveryControlFixture struct {
	recoveryStoreFixture
	sets      map[string]ports.RecoverySetRecord
	restores  map[string]ports.RestoreRecord
	identity  recoveryset.EnvironmentIdentity
	retention []ports.RetentionRecord
}

func (store *recoveryControlFixture) RequestRecoverySet(
	_ context.Context, record ports.RecoverySetRequestRecord,
) (int64, error) {
	if current, found := store.sets[record.RecoverySetID]; found {
		if current.OperationID == record.OperationID {
			return current.Sequence, nil
		}
		return 0, ports.ErrConflict
	}
	sequence := int64(len(store.sets) + 1)
	store.sets[record.RecoverySetID] = ports.RecoverySetRecord{
		RecoverySetID: record.RecoverySetID, OperationID: record.OperationID,
		EnvironmentID: record.EnvironmentID, Sequence: sequence,
		State: "requested", CreatedAt: record.CreatedAt, UpdatedAt: record.CreatedAt,
	}
	return sequence, nil
}

func (store *recoveryControlFixture) GetRecoverySet(
	_ context.Context, id string,
) (ports.RecoverySetRecord, error) {
	record, found := store.sets[id]
	if !found {
		return ports.RecoverySetRecord{}, ports.ErrNotFound
	}
	return record, nil
}

func (store *recoveryControlFixture) ListRecoverySets(
	_ context.Context, environmentID string,
) ([]ports.RecoverySetRecord, error) {
	result := make([]ports.RecoverySetRecord, 0, len(store.sets))
	for _, record := range store.sets {
		if record.EnvironmentID == environmentID {
			result = append(result, record)
		}
	}
	return result, nil
}

func (store *recoveryControlFixture) ClaimRecoveryManifest(
	_ context.Context, claim ports.RecoveryArtifactClaim,
) (ports.RecoveryArtifactRecord, error) {
	record, found := store.sets[claim.ArtifactID]
	if !found || len(record.ManifestBytes) == 0 {
		return ports.RecoveryArtifactRecord{}, ports.ErrNotFound
	}
	return ports.RecoveryArtifactRecord{SHA256: record.ManifestSHA256,
		Bytes: append([]byte(nil), record.ManifestBytes...)}, nil
}

func (store *recoveryControlFixture) ListRetentionCandidates(
	context.Context, string,
) ([]ports.RetentionRecord, error) {
	return append([]ports.RetentionRecord(nil), store.retention...), nil
}

func (store *recoveryControlFixture) MarkRecoverySetDeleting(
	_ context.Context, id, digest string, _ time.Time,
) error {
	for _, record := range store.retention {
		if record.RecoverySetID == id && record.ManifestSHA256 == digest && !record.LegalHold {
			return nil
		}
	}
	return ports.ErrConflict
}

func (store *recoveryControlFixture) RequestRestore(
	_ context.Context, request ports.RestoreRequestRecord,
) error {
	if _, found := store.restores[request.OperationID]; found {
		return ports.ErrConflict
	}
	store.restores[request.OperationID] = ports.RestoreRecord{
		OperationID: request.OperationID, RecoverySetID: request.RecoverySetID,
		EnvironmentID: request.EnvironmentID, Disposition: request.Disposition,
		State: "plan_ready", PlanSHA256: request.PlanSHA256,
		CreatedAt: request.CreatedAt, UpdatedAt: request.CreatedAt}
	return nil
}

func (store *recoveryControlFixture) ResolveRestore(
	_ context.Context, id string,
) (ports.RestoreRecord, error) {
	record, found := store.restores[id]
	if !found {
		return ports.RestoreRecord{}, ports.ErrNotFound
	}
	return record, nil
}

func (store *recoveryControlFixture) ClaimRestorePlan(
	context.Context, ports.RecoveryArtifactClaim,
) (ports.RecoveryArtifactRecord, error) {
	return ports.RecoveryArtifactRecord{}, ports.ErrNotFound
}

func (store *recoveryControlFixture) CurrentIdentity(
	context.Context, string,
) (ports.EnvironmentIdentityRecord, error) {
	return ports.EnvironmentIdentityRecord{Identity: store.identity,
		DatabaseSystemID: appUUID(8)}, nil
}
