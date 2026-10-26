package systemtransfer

import (
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type reconciliationRow struct {
	claimed                   bool
	actorID, originalJobID    *uuid.UUID
	token, authorization      *int64
	catalogVersionID          *uuid.UUID
	catalogSHA, queryDocument []byte
	querySHA                  []byte
	format                    *string
	originalGeneration        *int64
	originalEffectID          *uuid.UUID
	originalReceiptID         *uuid.UUID
	originalLeaseOwnerID      *uuid.UUID
	objectKey                 *string
	originalRequestSHA        []byte
	originalStartedAt         *time.Time
	keyID                     *string
	headerSHA, ciphertextSHA  []byte
	ciphertextBytes           *int64
	wrappedDEK, wrapNonce     []byte
	observationEffectID       *uuid.UUID
	observationReceiptID      *uuid.UUID
	observationRequestSHA     []byte
	observationStartedAt      *time.Time
}

func (raw *reconciliationRow) targets() []any {
	return []any{&raw.claimed, &raw.actorID, &raw.originalJobID, &raw.token,
		&raw.authorization, &raw.catalogVersionID, &raw.catalogSHA,
		&raw.queryDocument, &raw.querySHA, &raw.format, &raw.originalGeneration,
		&raw.originalEffectID, &raw.originalReceiptID, &raw.originalLeaseOwnerID,
		&raw.objectKey, &raw.originalRequestSHA, &raw.originalStartedAt,
		&raw.keyID, &raw.headerSHA, &raw.ciphertextSHA, &raw.ciphertextBytes,
		&raw.wrappedDEK, &raw.wrapNonce, &raw.observationEffectID,
		&raw.observationReceiptID, &raw.observationRequestSHA,
		&raw.observationStartedAt}
}

func (raw reconciliationRow) decode(
	claim app.ExportReconciliationClaim,
) (app.ExportReconciliationObservation, bool, error) {
	if !raw.claimed {
		return app.ExportReconciliationObservation{}, false, nil
	}
	if raw.actorID == nil || raw.originalJobID == nil || raw.token == nil ||
		raw.authorization == nil || raw.catalogVersionID == nil || raw.format == nil ||
		raw.originalGeneration == nil || raw.originalEffectID == nil ||
		raw.originalReceiptID == nil || raw.originalLeaseOwnerID == nil ||
		raw.objectKey == nil || raw.originalStartedAt == nil || raw.keyID == nil ||
		raw.ciphertextBytes == nil || raw.observationEffectID == nil ||
		raw.observationReceiptID == nil || raw.observationStartedAt == nil {
		return app.ExportReconciliationObservation{}, false,
			app.ErrExportStateConflict
	}
	catalog, catalogOK := workerDigest(raw.catalogSHA)
	query, queryOK := workerDigest(raw.querySHA)
	originalRequest, originalRequestOK := workerDigest(raw.originalRequestSHA)
	header, headerOK := workerDigest(raw.headerSHA)
	ciphertext, ciphertextOK := workerDigest(raw.ciphertextSHA)
	nonce, nonceOK := workerNonce(raw.wrapNonce)
	observationRequest, observationRequestOK := workerDigest(raw.observationRequestSHA)
	attempt := app.ExportAttempt{TenantID: claim.TenantID, ActorID: *raw.actorID,
		JobID: *raw.originalJobID, ExportID: claim.ExportID,
		AttemptID: *raw.originalEffectID, LeaseOwnerID: *raw.originalLeaseOwnerID,
		CatalogVersionID: *raw.catalogVersionID, Resource: claim.Resource,
		Format: *raw.format, State: app.ExportOutcomeUnknown,
		TokenVersion: *raw.token, AuthorizationVersion: *raw.authorization,
		CatalogSHA256: catalog, QuerySHA256: query,
		QueryDocument:     append([]byte(nil), raw.queryDocument...),
		AttemptGeneration: *raw.originalGeneration}
	draft := app.ExportArtifactDraftEvidence{KeyID: *raw.keyID,
		HeaderSHA256: header, CiphertextSHA256: ciphertext,
		CiphertextBytes: *raw.ciphertextBytes,
		WrappedDEK:      append([]byte(nil), raw.wrappedDEK...), WrapNonce: nonce}
	effect := app.ExportArtifactEffectStart{Attempt: attempt,
		EffectID: *raw.originalEffectID, EffectReceiptID: *raw.originalReceiptID,
		ObjectKey: *raw.objectKey, RequestSHA256: originalRequest,
		StartedAt: raw.originalStartedAt.UTC(), Draft: draft}
	observation := app.ExportReconciliationObservation{ExportEffect: effect,
		EffectID:        *raw.observationEffectID,
		EffectReceiptID: *raw.observationReceiptID,
		RequestSHA256:   observationRequest,
		StartedAt:       raw.observationStartedAt.UTC()}
	if !catalogOK || !queryOK || !originalRequestOK || !headerOK ||
		!ciphertextOK || !nonceOK || !observationRequestOK ||
		observation.Validate(claim) != nil {
		return app.ExportReconciliationObservation{}, false,
			app.ErrExportStateConflict
	}
	return observation, true, nil
}
