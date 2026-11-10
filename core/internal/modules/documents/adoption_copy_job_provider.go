package documents

import (
	"context"
	"errors"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var adoptionCopyReceiptNamespace = uuid.MustParse("91c4d2e8-5a17-4b6f-8d20-7e3a1c9b4f55")

type adoptionCopyProvider struct {
	identity effectexecution.Identity
	effect   AdoptionCopyEffect
	now      func() time.Time
}

func newAdoptionCopyProvider(effect AdoptionCopyEffect,
	now func() time.Time,
) (*adoptionCopyProvider, error) {
	identity, err := effectexecution.NewIdentity("documents.adoption", "copy")
	if err != nil || nilArchivePort(effect) || now == nil {
		return nil, ErrDocumentsRuntime
	}
	return &adoptionCopyProvider{identity: identity, effect: effect, now: now}, nil
}

func (provider *adoptionCopyProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

func (provider *adoptionCopyProvider) Execute(ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	parsed, jobID, err := provider.validateRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	err = provider.effect.Execute(withJobAuthority(ctx, request.Authority),
		foundation.ID(parsed.TenantID), foundation.ID(parsed.RunID),
		foundation.ID(parsed.SourceVersionID), jobID)
	if errors.Is(err, ErrAdoptionCopyUnavailable) {
		return provider.failed(request, "DOCUMENT_ADOPTION_COPY_UNAVAILABLE",
			effectsv1.ErrorPermanent)
	}
	return effectexecution.Unknown()
}

func (provider *adoptionCopyProvider) validateRequest(
	request effectsv1.Request,
) (jobcontract.AdoptionCopyReference, foundation.ID, error) {
	if provider == nil || request.Validate() != nil || request.TenantID == nil ||
		request.Authority.Domain != "documents" || request.Provider != "documents.adoption" ||
		request.Operation != "copy" || request.EffectRevision != 1 {
		return jobcontract.AdoptionCopyReference{}, foundation.ID{}, ErrDocumentsRuntime
	}
	parsed, err := jobcontract.ParseAdoptionCopy(request.PayloadReference)
	jobID, jobErr := foundation.ParseID(request.CausationID)
	_, digest, digestErr := jobcontract.AdoptionCopy(
		parsed.TenantID, parsed.RunID, parsed.SourceVersionID)
	if err != nil || jobErr != nil || digestErr != nil || !jobID.UUIDv7() ||
		request.EffectKind != "documents.adoption.copy" ||
		request.PayloadSHA256 != digest || *request.TenantID != parsed.TenantID.String() ||
		request.IdempotencyKey != jobcontract.AdoptionCopyEffectKey(
			parsed.RunID, parsed.SourceVersionID) {
		return jobcontract.AdoptionCopyReference{}, foundation.ID{}, ErrDocumentsRuntime
	}
	return parsed, jobID, nil
}

func (provider *adoptionCopyProvider) failed(request effectsv1.Request,
	code string, class effectsv1.ErrorClass,
) effectexecution.Result {
	reference := "documents.adoption"
	receipt, err := effectsv1.NewFailedReceipt(request,
		uuid.NewSHA1(adoptionCopyReceiptNamespace,
			[]byte(request.EffectID+"\x00"+code)).String(),
		code, class, &reference, provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

var _ effectexecution.Provider = (*adoptionCopyProvider)(nil)
