package systemtransfer

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (raw rawAttempt) decode(
	claim systemtransferapp.AttemptClaim,
) (systemtransferapp.AttemptState, error) {
	resource, resourceOK := publicresource.ParseResource(raw.resource)
	mapping, mappingErr := systemtransferapp.BuildStoredMappingProgram(resource, raw.mapping)
	header, headerOK := workerDigest(raw.header)
	ciphertext, ciphertextOK := workerDigest(raw.ciphertext)
	plaintext, plaintextOK := workerDigest(raw.plaintext)
	catalog, catalogOK := workerDigest(raw.catalogSHA)
	manifest, manifestOK := workerDigest(raw.manifestSHA)
	nonce, nonceOK := workerNonce(raw.wrapNonce)
	if !resourceOK || mappingErr != nil || !headerOK || !ciphertextOK ||
		!plaintextOK || !catalogOK || !manifestOK || !nonceOK {
		return systemtransferapp.AttemptState{}, systemtransferapp.ErrWorkerState
	}
	attempt := systemtransferapp.Attempt{
		ID: claim.ImportID, JobID: claim.JobID, AttemptID: claim.AttemptID,
		LeaseOwnerID: claim.LeaseOwnerID, LeaseGeneration: claim.LeaseGeneration,
		TenantID: claim.TenantID, ActorID: raw.actorID, Resource: resource,
		AuthorizationVersion: raw.authorization,
		CatalogVersionID:     raw.catalogVersionID,
		CatalogSHA256:        catalog, IdentityManifestSHA256: manifest,
		DuplicateAction: systemtransferapp.DuplicateAction(raw.duplicateAction),
		Mapping:         mapping,
	}
	source := systemtransferapp.ExactImportSource{
		TenantID: claim.TenantID, ActorID: raw.actorID, SourceID: raw.sourceID,
		Resource: resource, Format: raw.format,
		Artifact: systemtransferapp.SourceArtifact{
			ObjectKey: raw.objectKey, ObjectVersion: raw.objectVersion,
			HeaderSHA256: header, CiphertextSHA256: ciphertext,
			CiphertextBytes: raw.ciphertextBytes, PlaintextSHA256: plaintext,
			PlaintextBytes: raw.plaintextBytes, KeyID: raw.keyID,
			WrappedDEK: append([]byte(nil), raw.wrappedDEK...), WrapNonce: nonce,
			ProviderRequestID: raw.providerRequestID, UploadedAt: raw.uploadedAt.UTC(),
		},
	}
	state := systemtransferapp.AttemptState{Attempt: attempt, Source: source,
		NextRowIndex: raw.next, Processed: raw.processed, Finished: raw.finished}
	if state.Validate(claim) != nil {
		return systemtransferapp.AttemptState{}, systemtransferapp.ErrWorkerState
	}
	return state, nil
}
