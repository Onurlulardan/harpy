package dashboardexport

import (
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

type deleteClaimRow struct {
	disposition, state                     string
	content                                contentRow
	hasArtifact                            bool
	deleteActorID, deleteJobID, leaseOwner uuid.UUID
	deleteGeneration                       int64
}

func (stored *deleteClaimRow) destinations() []any {
	content := &stored.content
	return []any{&stored.disposition, &stored.state, &content.actorID, &content.exportID,
		&content.dashboardID, &content.versionID, &content.tokenVersion,
		&content.authorizationVersion, &content.authorizationClock, &content.manifest,
		&content.pinSHA256, &content.format, &content.parameters, &content.widgets,
		&content.maximumRows, &content.maximumBytes, &content.expiresAt,
		&content.generation, &stored.hasArtifact, &content.objectKey,
		&content.objectVersion, &content.keyID, &content.wrappedDEK, &content.wrapNonce,
		&content.headerSHA256, &content.ciphertextSHA256, &content.plaintextSHA256,
		&content.ciphertextBytes, &content.plaintextBytes, &content.rows,
		&content.contentType, &content.filename, &stored.deleteActorID,
		&stored.deleteJobID, &stored.deleteGeneration, &stored.leaseOwner}
}

func (stored deleteClaimRow) decode() (exportapp.DeleteClaim, error) {
	pin, err := stored.content.decodePin()
	if err != nil || !exportapp.State(stored.state).Valid() ||
		!validDisposition(stored.disposition) || stored.deleteActorID == uuid.Nil ||
		stored.deleteJobID.Version() != 7 {
		return exportapp.DeleteClaim{}, exportapp.ErrUnavailable
	}
	claim := exportapp.DeleteClaim{Pin: pin,
		Disposition: exportapp.ClaimDisposition(stored.disposition),
		State:       exportapp.State(stored.state), ObjectVersion: stored.content.objectVersion,
		DeleteActorID: stored.deleteActorID, DeleteJobID: stored.deleteJobID,
		LeaseOwner: stored.leaseOwner, DeleteGeneration: stored.deleteGeneration}
	if claim.Disposition == exportapp.ClaimTerminal {
		return claim, nil
	}
	if stored.leaseOwner == uuid.Nil || claim.DeleteGeneration < 1 ||
		claim.DeleteGeneration > exportapp.MaximumAttempts {
		return exportapp.DeleteClaim{}, exportapp.ErrUnavailable
	}
	if stored.hasArtifact {
		draft, draftErr := stored.content.decodeDraft(pin)
		if draftErr != nil {
			return exportapp.DeleteClaim{}, draftErr
		}
		claim.Draft = &draft
	}
	return claim, nil
}
