package notifications

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var deliveryReceiptNamespace = uuid.MustParse("a36d73f1-dd98-43c9-b074-c50442492789")

func (provider *deliveryProvider) succeeded(request effectsv1.Request,
	reference deliveryReference, code string,
) effectexecution.Result {
	receiptID := deliveryReceiptID(request, code)
	providerReference := reference.deliveryID.String()
	resultReference := "notification-delivery://receipt/" + reference.deliveryID.String()
	digest := sha256.Sum256([]byte(code + "\x00" + resultReference))
	receipt, err := effectsv1.NewSucceededReceipt(request, receiptID, code,
		&providerReference, resultReference, hex.EncodeToString(digest[:]),
		provider.now().UTC().Format(time.RFC3339Nano))
	return definitive(receipt, err)
}

func (provider *deliveryProvider) failed(request effectsv1.Request,
	reference deliveryReference, code string, class effectsv1.ErrorClass,
) effectexecution.Result {
	receiptID := deliveryReceiptID(request, code)
	providerReference := reference.deliveryID.String()
	receipt, err := effectsv1.NewFailedReceipt(request, receiptID, code, class,
		&providerReference, provider.now().UTC().Format(time.RFC3339Nano))
	return definitive(receipt, err)
}

func deliveryReceiptID(request effectsv1.Request, code string) string {
	return uuid.NewSHA1(deliveryReceiptNamespace,
		[]byte(request.EffectID+"\x00"+code)).String()
}

func definitive(receipt effectsv1.Receipt, err error) effectexecution.Result {
	if err != nil {
		return effectexecution.Unknown()
	}
	result, err := effectexecution.Definitive(receipt)
	if err != nil {
		return effectexecution.Unknown()
	}
	return result
}
