package documents

import (
	"context"
	"errors"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (provider *archiveProvider) Execute(ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	parsed, jobID, err := provider.validateRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	archiveID := foundation.ID(parsed.ResourceID)
	result, err := provider.effect.Execute(withJobAuthority(ctx, request.Authority),
		foundation.ID(parsed.TenantID), archiveID, jobID)
	if err == nil {
		return provider.succeeded(request, archiveID, result)
	}
	switch {
	case errors.Is(err, archiveapp.ErrLimit):
		return provider.failed(request, archiveID,
			"DOCUMENT_ARCHIVE_LIMIT_EXCEEDED", effectsv1.ErrorPermanent)
	case errors.Is(err, archiveapp.ErrForbidden):
		return provider.failed(request, archiveID,
			"DOCUMENT_ARCHIVE_AUTHORIZATION_REVOKED", effectsv1.ErrorPermanent)
	case errors.Is(err, archiveapp.ErrCancelled):
		return provider.failed(request, archiveID,
			"DOCUMENT_ARCHIVE_CANCELLED", effectsv1.ErrorPermanent)
	case errors.Is(err, archiveapp.ErrNotFound), errors.Is(err, archiveapp.ErrConflict):
		return provider.failed(request, archiveID,
			"DOCUMENT_ARCHIVE_STATE_CONFLICT", effectsv1.ErrorPermanent)
	default:
		return effectexecution.Unknown()
	}
}

func (provider *archiveProvider) validateRequest(
	request effectsv1.Request,
) (jobcontract.Reference, foundation.ID, error) {
	if provider == nil || request.Validate() != nil || request.TenantID == nil ||
		request.Authority.Domain != "documents" || request.Provider != "documents.archive" ||
		request.Operation != "build" || request.EffectRevision != 1 {
		return jobcontract.Reference{}, foundation.ID{}, ErrDocumentsRuntime
	}
	parsed, err := jobcontract.ParseArchive(request.PayloadReference)
	jobID, jobErr := foundation.ParseID(request.CausationID)
	_, digest, digestErr := jobcontract.Archive(parsed.TenantID, parsed.ResourceID)
	if err != nil || jobErr != nil || digestErr != nil || !jobID.UUIDv7() ||
		request.EffectKind != "documents.archive.build" ||
		request.PayloadSHA256 != digest || *request.TenantID != parsed.TenantID.String() ||
		request.IdempotencyKey != "documents:archive:"+parsed.ResourceID.String() {
		return jobcontract.Reference{}, foundation.ID{}, ErrDocumentsRuntime
	}
	return parsed, jobID, nil
}
