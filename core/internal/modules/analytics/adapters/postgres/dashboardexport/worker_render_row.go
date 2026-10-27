package dashboardexport

import (
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

type renderClaimRow struct {
	disposition, state string
	content            contentRow
	jobID, leaseOwner  uuid.UUID
	effectStarted      bool
}

func (stored *renderClaimRow) destinations() []any {
	content := &stored.content
	return []any{&stored.disposition, &stored.state, &content.actorID, &content.exportID,
		&content.dashboardID, &content.versionID, &stored.jobID, &content.tokenVersion,
		&content.authorizationVersion, &content.authorizationClock, &content.manifest,
		&content.pinSHA256, &content.format, &content.parameters, &content.widgets,
		&content.maximumRows, &content.maximumBytes, &content.expiresAt,
		&content.generation, &stored.leaseOwner, &stored.effectStarted,
		&content.objectKey, &content.keyID, &content.wrappedDEK, &content.wrapNonce,
		&content.headerSHA256, &content.ciphertextSHA256, &content.plaintextSHA256,
		&content.ciphertextBytes, &content.plaintextBytes, &content.rows,
		&content.contentType, &content.filename}
}

func (stored renderClaimRow) decode() (exportapp.RenderClaim, error) {
	pin, err := stored.content.decodePin()
	if err != nil || !exportapp.State(stored.state).Valid() || stored.jobID.Version() != 7 ||
		!validDisposition(stored.disposition) {
		return exportapp.RenderClaim{}, exportapp.ErrUnavailable
	}
	claim := exportapp.RenderClaim{Disposition: exportapp.ClaimDisposition(stored.disposition),
		Pin: pin, JobID: stored.jobID, LeaseOwner: stored.leaseOwner,
		Generation: stored.content.generation, State: exportapp.State(stored.state)}
	if claim.Disposition == exportapp.ClaimTerminal {
		return claim, nil
	}
	if stored.leaseOwner == uuid.Nil || claim.Generation < 1 ||
		claim.Generation > exportapp.MaximumAttempts ||
		(claim.Disposition == exportapp.ClaimExecute && stored.effectStarted) ||
		(claim.Disposition == exportapp.ClaimReconcile && !stored.effectStarted) {
		return exportapp.RenderClaim{}, exportapp.ErrUnavailable
	}
	if stored.effectStarted {
		draft, draftErr := stored.content.decodeDraft(pin)
		if draftErr != nil {
			return exportapp.RenderClaim{}, draftErr
		}
		claim.Draft = &draft
	}
	return claim, nil
}

func validDisposition(value string) bool {
	return value == string(exportapp.ClaimExecute) ||
		value == string(exportapp.ClaimReconcile) || value == string(exportapp.ClaimTerminal)
}
