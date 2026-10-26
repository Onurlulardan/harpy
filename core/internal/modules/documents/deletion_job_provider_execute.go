package documents

import (
	"context"
	"errors"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (provider *deletionProvider) Execute(ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	parsed, jobID, err := provider.validateRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	deletionID := foundation.ID(parsed.ResourceID)
	result, err := provider.effect.Execute(withJobAuthority(ctx, request.Authority),
		foundation.ID(parsed.TenantID), deletionID, jobID)
	if err == nil {
		return provider.succeeded(request, deletionID, result)
	}
	switch {
	case errors.Is(err, deletionapp.ErrNotFound):
		return provider.failed(request, deletionID,
			"DOCUMENT_BLOB_DELETION_NOT_FOUND", effectsv1.ErrorPermanent)
	case errors.Is(err, deletionapp.ErrConflict):
		return provider.failed(request, deletionID,
			"DOCUMENT_BLOB_DELETION_CONFLICT", effectsv1.ErrorConflict)
	default:
		return effectexecution.Unknown()
	}
}

func (provider *deletionProvider) validateRequest(
	request effectsv1.Request,
) (jobcontract.Reference, foundation.ID, error) {
	if provider == nil || request.Validate() != nil || request.TenantID == nil ||
		request.Authority.Domain != "documents" || request.Provider != "documents.blob" ||
		request.Operation != "delete" || request.EffectRevision != 1 {
		return jobcontract.Reference{}, foundation.ID{}, ErrDocumentsRuntime
	}
	parsed, err := jobcontract.ParseDeletion(request.PayloadReference)
	jobID, jobErr := foundation.ParseID(request.CausationID)
	_, digest, digestErr := jobcontract.Deletion(parsed.TenantID, parsed.ResourceID)
	if err != nil || jobErr != nil || digestErr != nil || !jobID.UUIDv7() ||
		request.EffectKind != "documents.blob.delete" ||
		request.PayloadSHA256 != digest || *request.TenantID != parsed.TenantID.String() ||
		request.IdempotencyKey != "documents:blob-delete:"+parsed.ResourceID.String() {
		return jobcontract.Reference{}, foundation.ID{}, ErrDocumentsRuntime
	}
	return parsed, jobID, nil
}
