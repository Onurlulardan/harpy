package iam

import (
	"encoding/hex"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var lifecycleDeliveryReceiptNamespace = uuid.MustParse(
	"af0122f9-f239-49aa-a947-6435e282c0d5")

func (provider *lifecycleDeliveryProvider) receiptForStored(
	request lifecycleDeliveryRequest,
	stored lifecycledeliveryapp.StoredResult) effectexecution.Result {
	if stored.Validate(request.tenantID) != nil ||
		stored.EffectKey != request.reference.EffectKey ||
		stored.MessageID != request.messageID {
		return effectexecution.Unknown()
	}
	providerReference := stored.ProviderReference
	receiptID := uuid.NewSHA1(lifecycleDeliveryReceiptNamespace,
		[]byte(request.request.EffectID+"\x00"+stored.ReceiptID.String())).String()
	completedAt := canonicalCompletion(stored.CompletedAt, request.requestedAt)
	if stored.Outcome == lifecycledeliveryapp.OutcomeDelivered {
		resultReference := "iam-lifecycle-delivery://receipt/" + stored.ReceiptID.String()
		receipt, err := effectsv1.NewSucceededReceipt(request.request, receiptID,
			stored.Code, &providerReference, resultReference,
			hex.EncodeToString(stored.ResultSHA256[:]), completedAt)
		return definitiveOrUnknown(receipt, err)
	}
	receipt, err := effectsv1.NewFailedReceipt(request.request, receiptID,
		stored.Code, effectsv1.ErrorPermanent, &providerReference, completedAt)
	return definitiveOrUnknown(receipt, err)
}

func (provider *lifecycleDeliveryProvider) failed(request lifecycleDeliveryRequest,
	code string, class effectsv1.ErrorClass) effectexecution.Result {
	providerReference := request.messageID
	receiptID := uuid.NewSHA1(lifecycleDeliveryReceiptNamespace,
		[]byte(request.request.EffectID+"\x00"+code)).String()
	receipt, err := effectsv1.NewFailedReceipt(request.request, receiptID,
		code, class, &providerReference,
		canonicalCompletion(provider.now().UTC(), request.requestedAt))
	return definitiveOrUnknown(receipt, err)
}
