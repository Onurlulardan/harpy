package iam

import (
	"context"
	"crypto/sha256"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

const lifecycleDeliveryFinalizeTimeout = 5 * time.Second

var lifecycleDeliveryStoredReceiptNamespace = uuid.MustParse(
	"5e5590c9-704a-44a8-aa89-29cbe941d98b")

func (provider *lifecycleDeliveryProvider) releaseThenFailed(ctx context.Context,
	request lifecycleDeliveryRequest, code string) effectexecution.Result {
	bounded, cancel := provider.finalizationContext(ctx)
	defer cancel()
	if err := provider.state.Release(bounded, request.reference); err != nil {
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
	return provider.failed(request, code, effectsv1.ErrorTransient)
}

func (provider *lifecycleDeliveryProvider) finalizePermanent(ctx context.Context,
	request lifecycleDeliveryRequest, code string,
	digest [32]byte) effectexecution.Result {
	return provider.finalize(ctx, request, lifecycledeliveryapp.OutcomePermanent,
		code, digest, provider.now().UTC())
}

func (provider *lifecycleDeliveryProvider) finalize(ctx context.Context,
	request lifecycleDeliveryRequest, outcome lifecycledeliveryapp.Outcome,
	code string, digest [32]byte, completedAt time.Time) effectexecution.Result {
	result := lifecycledeliveryapp.StoredResult{
		ReceiptID: uuid.NewSHA1(lifecycleDeliveryStoredReceiptNamespace,
			[]byte(request.reference.EffectID.String()+"\x00"+code)),
		ArtifactRef: request.reference.ArtifactRef, EffectKey: request.reference.EffectKey,
		MessageID: request.messageID, Outcome: outcome, Code: code,
		ProviderReference: request.messageID, ResultSHA256: digest,
		CompletedAt: completedAt,
	}
	bounded, cancel := provider.finalizationContext(ctx)
	defer cancel()
	stored, err := provider.state.Finalize(bounded, lifecycledeliveryapp.FinalizeCommand{
		EffectReference: request.reference, Result: result,
	})
	if err != nil {
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
	return provider.receiptForStored(request, stored)
}

func (provider *lifecycleDeliveryProvider) handleSMTPResult(ctx context.Context,
	request lifecycleDeliveryRequest, result smtp.Result) effectexecution.Result {
	if result.ProviderReference != request.messageID || result.EvidenceSHA256 == [32]byte{} ||
		result.CompletedAt.IsZero() || result.CompletedAt.Before(request.requestedAt) {
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
	switch result.Kind {
	case smtp.ResultAccepted:
		return provider.finalize(ctx, request, lifecycledeliveryapp.OutcomeDelivered,
			result.Code, result.EvidenceSHA256, result.CompletedAt.UTC())
	case smtp.ResultPermanent:
		return provider.finalize(ctx, request, lifecycledeliveryapp.OutcomePermanent,
			result.Code, result.EvidenceSHA256, result.CompletedAt.UTC())
	case smtp.ResultTransient:
		return provider.releaseThenFailed(ctx, request, result.Code)
	default:
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
}

func (provider *lifecycleDeliveryProvider) persistUnknown(ctx context.Context,
	reference lifecycledeliveryapp.EffectReference) {
	reference.OccurredAt = provider.now().UTC()
	bounded, cancel := provider.finalizationContext(ctx)
	defer cancel()
	_ = provider.state.MarkUnknown(bounded, reference)
}

func (provider *lifecycleDeliveryProvider) finalizationContext(
	ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(context.WithoutCancel(ctx), lifecycleDeliveryFinalizeTimeout)
}

func lifecycleDeliveryFailureDigest(code, artifactRef string) [32]byte {
	return sha256.Sum256([]byte(
		"iam.lifecycle.delivery.failure.v1\x00" + code + "\x00" + artifactRef))
}
