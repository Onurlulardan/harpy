package iam

import (
	"encoding/hex"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var invitationDeliveryReceiptNamespace = uuid.MustParse("f6c29bcc-440d-4c30-b86f-e1e1f9b145d5")

func (provider *invitationDeliveryProvider) receiptForStored(
	request invitationRequest,
	stored invitationdeliveryapp.StoredResult,
) effectexecution.Result {
	if stored.Validate(request.tenantID) != nil ||
		stored.EffectKey != request.reference.EffectKey ||
		stored.MessageID != request.messageID {
		return effectexecution.Unknown()
	}
	providerReference := stored.ProviderReference
	receiptID := uuid.NewSHA1(
		invitationDeliveryReceiptNamespace,
		[]byte(request.request.EffectID+"\x00"+stored.ReceiptID.String()),
	).String()
	completedAt := canonicalCompletion(stored.CompletedAt, request.requestedAt)
	if stored.Outcome == invitationdeliveryapp.OutcomeDelivered {
		resultReference := "iam-invitation-delivery://receipt/" + stored.ReceiptID.String()
		receipt, err := effectsv1.NewSucceededReceipt(
			request.request, receiptID, stored.Code, &providerReference,
			resultReference, hex.EncodeToString(stored.ResultSHA256[:]), completedAt,
		)
		return definitiveOrUnknown(receipt, err)
	}
	receipt, err := effectsv1.NewFailedReceipt(
		request.request, receiptID, stored.Code, effectsv1.ErrorPermanent,
		&providerReference, completedAt,
	)
	return definitiveOrUnknown(receipt, err)
}

func (provider *invitationDeliveryProvider) failed(
	request invitationRequest,
	code string,
	class effectsv1.ErrorClass,
) effectexecution.Result {
	providerReference := request.messageID
	receiptID := uuid.NewSHA1(
		invitationDeliveryReceiptNamespace,
		[]byte(request.request.EffectID+"\x00"+code),
	).String()
	receipt, err := effectsv1.NewFailedReceipt(
		request.request, receiptID, code, class, &providerReference,
		canonicalCompletion(provider.now().UTC(), request.requestedAt),
	)
	return definitiveOrUnknown(receipt, err)
}

func definitiveOrUnknown(
	receipt effectsv1.Receipt,
	err error,
) effectexecution.Result {
	if err != nil {
		return effectexecution.Unknown()
	}
	result, err := effectexecution.Definitive(receipt)
	if err != nil {
		return effectexecution.Unknown()
	}
	return result
}

func canonicalCompletion(completedAt time.Time, requestedAt time.Time) string {
	completedAt = completedAt.UTC()
	if completedAt.Before(requestedAt) {
		completedAt = requestedAt.UTC()
	}
	return completedAt.Format(time.RFC3339Nano)
}
