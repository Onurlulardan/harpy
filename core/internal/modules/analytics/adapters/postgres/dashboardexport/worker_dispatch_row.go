package dashboardexport

import (
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

type dispatchClaimRow struct {
	claimState, disposition, state, format, objectKey, keyID     *string
	contentType, filename                                        *string
	attemptID, actorID, exportID, dashboardID, versionID, jobID  *uuid.UUID
	leaseOwner                                                   *uuid.UUID
	claimedAt, leaseUntil, expiresAt                             *time.Time
	tokenVersion, authorizationVersion, authorizationClock       *int64
	maximumRows, maximumBytes, generation                        *int64
	ciphertextBytes, plaintextBytes, rows                        *int64
	effectStarted                                                *bool
	manifest, pinSHA, parameters, widgets, wrappedDEK, wrapNonce []byte
	headerSHA, ciphertextSHA, plaintextSHA                       []byte
}

func (stored *dispatchClaimRow) targets() []any {
	return []any{&stored.claimState, &stored.attemptID, &stored.claimedAt,
		&stored.leaseUntil, &stored.disposition, &stored.state, &stored.actorID,
		&stored.exportID, &stored.dashboardID, &stored.versionID, &stored.jobID,
		&stored.tokenVersion, &stored.authorizationVersion,
		&stored.authorizationClock, &stored.manifest, &stored.pinSHA,
		&stored.format, &stored.parameters, &stored.widgets, &stored.maximumRows,
		&stored.maximumBytes, &stored.expiresAt, &stored.generation,
		&stored.leaseOwner, &stored.effectStarted, &stored.objectKey, &stored.keyID,
		&stored.wrappedDEK, &stored.wrapNonce, &stored.headerSHA,
		&stored.ciphertextSHA, &stored.plaintextSHA, &stored.ciphertextBytes,
		&stored.plaintextBytes, &stored.rows, &stored.contentType, &stored.filename}
}

func (stored dispatchClaimRow) decode(candidate exportapp.RenderCandidate,
	reservation exportapp.SpoolReservationIdentity,
) (exportapp.RenderDispatchAttempt, error) {
	if stored.claimState == nil {
		return exportapp.RenderDispatchAttempt{}, exportapp.ErrUnavailable
	}
	if *stored.claimState == "quota_wait" {
		return exportapp.RenderDispatchAttempt{}, exportapp.ErrDispatchQuotaWait
	}
	if *stored.claimState != "claimed" || !stored.complete() {
		return exportapp.RenderDispatchAttempt{}, exportapp.ErrUnavailable
	}
	content := contentRow{tenantID: candidate.TenantID, actorID: *stored.actorID,
		exportID: *stored.exportID, dashboardID: *stored.dashboardID,
		versionID: *stored.versionID, tokenVersion: *stored.tokenVersion,
		authorizationVersion: *stored.authorizationVersion,
		authorizationClock:   *stored.authorizationClock, manifest: stored.manifest,
		pinSHA256: stored.pinSHA, format: *stored.format, parameters: stored.parameters,
		widgets: stored.widgets, maximumRows: *stored.maximumRows,
		maximumBytes: *stored.maximumBytes, expiresAt: *stored.expiresAt,
		generation: *stored.generation, objectKey: *stored.objectKey, keyID: *stored.keyID,
		wrappedDEK: stored.wrappedDEK, wrapNonce: stored.wrapNonce,
		headerSHA256: stored.headerSHA, ciphertextSHA256: stored.ciphertextSHA,
		plaintextSHA256: stored.plaintextSHA, ciphertextBytes: *stored.ciphertextBytes,
		plaintextBytes: *stored.plaintextBytes, rows: *stored.rows,
		contentType: *stored.contentType, filename: *stored.filename}
	claimRow := renderClaimRow{disposition: *stored.disposition, state: *stored.state,
		content: content, jobID: *stored.jobID, leaseOwner: *stored.leaseOwner,
		effectStarted: *stored.effectStarted}
	claim, err := claimRow.decode()
	if err != nil || claim.Disposition != exportapp.ClaimExecute ||
		*stored.exportID != candidate.ExportID || *stored.jobID != candidate.JobID ||
		*stored.generation != candidate.AttemptGeneration ||
		*stored.leaseOwner != reservation.LeaseOwnerID ||
		stored.attemptID.Version() != 7 || stored.claimedAt.IsZero() ||
		!stored.leaseUntil.After(*stored.claimedAt) {
		return exportapp.RenderDispatchAttempt{}, exportapp.ErrUnavailable
	}
	return exportapp.RenderDispatchAttempt{Candidate: candidate, Claim: claim,
		AttemptID: *stored.attemptID, ClaimedAt: stored.claimedAt.UTC(),
		LeaseUntil: stored.leaseUntil.UTC()}, nil
}

func (stored dispatchClaimRow) complete() bool {
	return stored.attemptID != nil && stored.claimedAt != nil && stored.leaseUntil != nil &&
		stored.disposition != nil && stored.state != nil && stored.actorID != nil &&
		stored.exportID != nil && stored.dashboardID != nil && stored.versionID != nil &&
		stored.jobID != nil && stored.tokenVersion != nil && stored.authorizationVersion != nil &&
		stored.authorizationClock != nil && stored.format != nil && stored.maximumRows != nil &&
		stored.maximumBytes != nil && stored.expiresAt != nil && stored.generation != nil &&
		stored.leaseOwner != nil && stored.effectStarted != nil && stored.objectKey != nil &&
		stored.keyID != nil && stored.ciphertextBytes != nil && stored.plaintextBytes != nil &&
		stored.rows != nil && stored.contentType != nil && stored.filename != nil
}
