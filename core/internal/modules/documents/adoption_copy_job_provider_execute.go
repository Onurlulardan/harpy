package documents

import (
	"context"
	"errors"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (provider *adoptionCopyProvider) Execute(ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	parsed, authority, err := provider.validateRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	result, err := provider.effect.Execute(withJobAuthority(ctx, request.Authority), authority,
		foundation.ID(parsed.SourceVersionID))
	if err == nil {
		return provider.succeeded(request, parsed, result)
	}
	switch {
	case errors.Is(err, adoptionapp.ErrBlocked):
		return provider.failed(request, parsed, "DOCUMENT_ADOPTION_COPY_BLOCKED",
			effectsv1.ErrorPermanent)
	case errors.Is(err, adoptionapp.ErrConflict), errors.Is(err, adoptionapp.ErrInvalid):
		return provider.failed(request, parsed, "DOCUMENT_ADOPTION_COPY_CONFLICT",
			effectsv1.ErrorPermanent)
	default:
		return effectexecution.Unknown()
	}
}

func (provider *adoptionCopyProvider) validateRequest(
	request effectsv1.Request,
) (jobcontract.AdoptionCopyReference, adoptionapp.Authority, error) {
	if provider == nil || request.Validate() != nil || request.TenantID == nil ||
		request.Authority.Domain != "documents" ||
		request.Provider != "documents.adoption" || request.Operation != "copy" ||
		request.EffectRevision != 1 {
		return jobcontract.AdoptionCopyReference{}, adoptionapp.Authority{},
			ErrDocumentsRuntime
	}
	parsed, err := jobcontract.ParseAdoptionCopy(request.PayloadReference)
	_, digest, digestErr := jobcontract.AdoptionCopy(parsed.TenantID,
		parsed.ResourceID, parsed.SourceVersionID, parsed.CutoverGeneration,
		parsed.ManifestSHA256)
	if err != nil || digestErr != nil ||
		request.EffectKind != "documents.adoption.copy" ||
		request.PayloadSHA256 != digest || *request.TenantID != parsed.TenantID.String() ||
		request.Authority.CutoverGeneration != int64(parsed.CutoverGeneration) ||
		request.IdempotencyKey != adoptionCopyKey(parsed.ResourceID.String(),
			parsed.SourceVersionID.String()) {
		return jobcontract.AdoptionCopyReference{}, adoptionapp.Authority{},
			ErrDocumentsRuntime
	}
	authority := adoptionapp.Authority{TenantID: foundation.ID(parsed.TenantID),
		RunID: foundation.ID(parsed.ResourceID), CutoverGeneration: parsed.CutoverGeneration,
		ManifestSHA256: parsed.ManifestSHA256}
	return parsed, authority, nil
}
