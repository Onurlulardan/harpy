package documents

import (
	"context"
	"errors"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (provider *uploadReconcileProvider) Execute(ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	parsed, err := provider.validateRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	result, err := provider.effect.Execute(withJobAuthority(ctx, request.Authority),
		foundation.ID(parsed.TenantID),
		foundation.ID(parsed.ResourceID), parsed.Generation)
	if err == nil {
		return provider.succeeded(request, parsed, result)
	}
	if errors.Is(err, uploadreconcileapp.ErrConflict) {
		return provider.failed(request, foundation.ID(parsed.ResourceID),
			"DOCUMENT_UPLOAD_RECONCILE_NOT_DUE", effectsv1.ErrorConflict)
	}
	return effectexecution.Unknown()
}

func (provider *uploadReconcileProvider) validateRequest(
	request effectsv1.Request,
) (jobcontract.UploadReference, error) {
	if provider == nil || request.Validate() != nil || request.TenantID == nil ||
		request.Authority.Domain != "documents" || request.Provider != "documents.upload" ||
		request.Operation != "reconcile" || request.EffectRevision != 1 ||
		request.EffectKind != "documents.upload.reconcile" {
		return jobcontract.UploadReference{}, ErrDocumentsRuntime
	}
	parsed, err := jobcontract.ParseUploadReconcile(request.PayloadReference)
	_, digest, digestErr := jobcontract.UploadReconcile(parsed.TenantID,
		parsed.ResourceID, parsed.Generation)
	if err != nil || digestErr != nil || request.PayloadSHA256 != digest ||
		*request.TenantID != parsed.TenantID.String() ||
		request.IdempotencyKey != uploadReconcileKey(parsed.ResourceID.String(),
			parsed.Generation) {
		return jobcontract.UploadReference{}, foundation.ErrInvalidID
	}
	return parsed, nil
}
