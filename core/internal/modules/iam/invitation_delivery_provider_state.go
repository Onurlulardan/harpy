package iam

import (
	"context"
	"crypto/sha256"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

const invitationDeliveryFinalizeTimeout = 5 * time.Second

var invitationDeliveryStoredReceiptNamespace = uuid.MustParse("b46d987f-aadf-4cc5-8d20-630a84e23f54")

func (provider *invitationDeliveryProvider) releaseThenFailed(
	ctx context.Context,
	request invitationRequest,
	code string,
) effectexecution.Result {
	bounded, cancel := provider.finalizationContext(ctx)
	defer cancel()
	if err := provider.state.Release(bounded, request.reference); err != nil {
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
	return provider.failed(request, code, effectsv1.ErrorTransient)
}

func (provider *invitationDeliveryProvider) finalizePermanent(
	ctx context.Context,
	request invitationRequest,
	code string,
	digest [32]byte,
) effectexecution.Result {
	return provider.finalize(
		ctx, request, invitationdeliveryapp.OutcomePermanent,
		code, digest, provider.now().UTC(),
	)
}

func (provider *invitationDeliveryProvider) finalize(
	ctx context.Context,
	request invitationRequest,
	outcome invitationdeliveryapp.Outcome,
	code string,
	digest [32]byte,
	completedAt time.Time,
) effectexecution.Result {
	result := invitationdeliveryapp.StoredResult{
		ReceiptID: uuid.NewSHA1(
			invitationDeliveryStoredReceiptNamespace,
			[]byte(request.reference.EffectID.String()+"\x00"+code),
		),
		ArtifactRef:       request.reference.ArtifactRef,
		EffectKey:         request.reference.EffectKey,
		MessageID:         request.messageID,
		Outcome:           outcome,
		Code:              code,
		ProviderReference: request.messageID,
		ResultSHA256:      digest,
		CompletedAt:       completedAt,
	}
	bounded, cancel := provider.finalizationContext(ctx)
	defer cancel()
	stored, err := provider.state.Finalize(bounded, invitationdeliveryapp.FinalizeCommand{
		EffectReference: request.reference,
		Result:          result,
	})
	if err != nil {
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
	return provider.receiptForStored(request, stored)
}

func (provider *invitationDeliveryProvider) handleSMTPResult(
	ctx context.Context,
	request invitationRequest,
	result smtp.Result,
) effectexecution.Result {
	if result.ProviderReference != request.messageID || result.EvidenceSHA256 == [32]byte{} ||
		result.CompletedAt.IsZero() || result.CompletedAt.Before(request.requestedAt) {
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
	switch result.Kind {
	case smtp.ResultAccepted:
		return provider.finalize(
			ctx, request, invitationdeliveryapp.OutcomeDelivered,
			result.Code, result.EvidenceSHA256, result.CompletedAt.UTC(),
		)
	case smtp.ResultPermanent:
		return provider.finalize(
			ctx, request, invitationdeliveryapp.OutcomePermanent,
			result.Code, result.EvidenceSHA256, result.CompletedAt.UTC(),
		)
	case smtp.ResultTransient:
		return provider.releaseThenFailed(ctx, request, result.Code)
	default:
		provider.persistUnknown(ctx, request.reference)
		return effectexecution.Unknown()
	}
}

func (provider *invitationDeliveryProvider) persistUnknown(
	ctx context.Context,
	reference invitationdeliveryapp.EffectReference,
) {
	reference.OccurredAt = provider.now().UTC()
	bounded, cancel := provider.finalizationContext(ctx)
	defer cancel()
	_ = provider.state.MarkUnknown(bounded, reference)
}

func (provider *invitationDeliveryProvider) finalizationContext(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(context.WithoutCancel(ctx), invitationDeliveryFinalizeTimeout)
}

func deliveryFailureDigest(code string, artifactRef string) [32]byte {
	return sha256.Sum256([]byte("iam.invitation.delivery.failure.v1\x00" + code + "\x00" + artifactRef))
}
