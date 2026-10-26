package dashboardexport

import (
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

type maintenanceClaimRow struct {
	disposition                                      string
	generation                                       *int64
	leaseUntil, localStateAt, deleteVerifiedAt       *time.Time
	mode, state, format, objectKey, objectVersion    *string
	keyID, contentType, filename, deleteProviderID   *string
	wrappedDEK, wrapNonce, headerSHA, ciphertextSHA  []byte
	plaintextSHA, localStateSHA, deleteResponseSHA   []byte
	deleteAbsenceSHA                                 []byte
	ciphertextBytes, plaintextBytes, rows, renderGen *int64
	deleteAttemptGeneration                          *int64
}

func (stored *maintenanceClaimRow) targets() []any {
	return []any{&stored.disposition, &stored.generation, &stored.leaseUntil,
		&stored.mode, &stored.state, &stored.format, &stored.objectKey,
		&stored.objectVersion, &stored.keyID, &stored.wrappedDEK, &stored.wrapNonce,
		&stored.headerSHA, &stored.ciphertextSHA, &stored.plaintextSHA,
		&stored.ciphertextBytes, &stored.plaintextBytes, &stored.rows,
		&stored.contentType, &stored.filename, &stored.renderGen,
		&stored.deleteAttemptGeneration,
		&stored.localStateSHA, &stored.localStateAt, &stored.deleteProviderID,
		&stored.deleteResponseSHA, &stored.deleteAbsenceSHA, &stored.deleteVerifiedAt}
}

func (stored maintenanceClaimRow) decode(candidate exportapp.MaintenanceUnknownCandidate,
	cleanupOwner uuid.UUID,
) (exportapp.MaintenanceUnknownClaim, error) {
	claim := exportapp.MaintenanceUnknownClaim{Candidate: candidate,
		Disposition:         exportapp.UnknownClaimDisposition(stored.disposition),
		CleanupLeaseOwnerID: cleanupOwner}
	if claim.Disposition == exportapp.UnknownBusy ||
		claim.Disposition == exportapp.UnknownResolved {
		if stored.generation != nil || stored.leaseUntil != nil {
			return exportapp.MaintenanceUnknownClaim{}, exportapp.ErrUnavailable
		}
		return claim, nil
	}
	if claim.Disposition != exportapp.UnknownClaimed || stored.generation == nil ||
		stored.leaseUntil == nil || stored.mode == nil || stored.state == nil ||
		stored.localStateAt == nil || len(stored.localStateSHA) != 32 ||
		stored.deleteAttemptGeneration == nil ||
		*stored.generation < 1 || !stored.leaseUntil.After(candidate.UnknownAt) {
		return exportapp.MaintenanceUnknownClaim{}, exportapp.ErrUnavailable
	}
	claim.CleanupGeneration = *stored.generation
	claim.DeleteAttemptGeneration = *stored.deleteAttemptGeneration
	claim.CleanupLeaseUntil = stored.leaseUntil.UTC()
	claim.Mode = exportapp.MaintenanceCleanupMode(*stored.mode)
	claim.TerminalState = exportapp.State(*stored.state)
	if !claim.TerminalState.Valid() || !validMaintenanceTerminal(candidate.Kind,
		claim.TerminalState) {
		return exportapp.MaintenanceUnknownClaim{}, exportapp.ErrUnavailable
	}
	copy(claim.StoredEvidence.LocalStateSHA256[:], stored.localStateSHA)
	claim.StoredEvidence.LocalStateAt = stored.localStateAt.UTC()
	stored.copyDeleteEvidence(&claim.StoredEvidence)
	if claim.Mode == exportapp.MaintenanceCleanupArtifact {
		if claim.DeleteAttemptGeneration < 1 || claim.DeleteAttemptGeneration > 3 ||
			stored.renderGen == nil || stored.objectVersion == nil ||
			*stored.objectVersion == "" {
			return exportapp.MaintenanceUnknownClaim{}, exportapp.ErrUnavailable
		}
		draftRow := recoveryClaimRow{format: stored.format, objectKey: stored.objectKey,
			keyID: stored.keyID, wrappedDEK: stored.wrappedDEK,
			wrapNonce: stored.wrapNonce, headerSHA: stored.headerSHA,
			ciphertextSHA: stored.ciphertextSHA, plaintextSHA: stored.plaintextSHA,
			ciphertextBytes: stored.ciphertextBytes,
			plaintextBytes:  stored.plaintextBytes, rows: stored.rows,
			contentType: stored.contentType, filename: stored.filename}
		draft, err := draftRow.draft(*stored.renderGen)
		if err != nil {
			return exportapp.MaintenanceUnknownClaim{}, err
		}
		claim.Draft, claim.ObjectVersion = &draft, *stored.objectVersion
	} else if claim.Mode != exportapp.MaintenanceCleanupTerminal &&
		claim.Mode != exportapp.MaintenanceCleanupCompleted {
		return exportapp.MaintenanceUnknownClaim{}, exportapp.ErrUnavailable
	} else if claim.DeleteAttemptGeneration < 0 || claim.DeleteAttemptGeneration > 3 {
		return exportapp.MaintenanceUnknownClaim{}, exportapp.ErrUnavailable
	}
	return claim, nil
}

func (stored maintenanceClaimRow) copyDeleteEvidence(
	evidence *exportapp.MaintenanceResolutionEvidence,
) {
	if stored.objectKey != nil {
		evidence.ObjectKey = *stored.objectKey
	}
	if stored.objectVersion != nil {
		evidence.ObjectVersion = *stored.objectVersion
	}
	if stored.deleteProviderID != nil {
		evidence.DeleteProviderRequestID = *stored.deleteProviderID
	}
	copy(evidence.DeleteResponseSHA256[:], stored.deleteResponseSHA)
	copy(evidence.AbsenceSHA256[:], stored.deleteAbsenceSHA)
	if stored.deleteVerifiedAt != nil {
		evidence.DeletedAt = stored.deleteVerifiedAt.UTC()
	}
}

func validMaintenanceTerminal(kind string, state exportapp.State) bool {
	if kind == "analytics.dashboard.export.delete" {
		return state == exportapp.StateDeleted
	}
	return state == exportapp.StateExpired || state == exportapp.StateCancelled ||
		state == exportapp.StateFailed || state == exportapp.StateDeleted
}
