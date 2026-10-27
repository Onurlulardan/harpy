package documents

import (
	"encoding/hex"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var deletionReceiptNamespace = uuid.MustParse("4c39cb86-28e2-414b-adc2-b649f2b903a3")

func (provider *deletionProvider) succeeded(request effectsv1.Request,
	deletionID foundation.ID, result deletionapp.Result,
) effectexecution.Result {
	if result.ID != deletionID || result.TenantID.String() != *request.TenantID ||
		result.ProviderRequestID == "" || result.AbsenceSHA256 == [32]byte{} ||
		result.DeleteSHA256 == [32]byte{} || result.ConfirmedAt.IsZero() {
		return effectexecution.Unknown()
	}
	code := "DOCUMENT_BLOB_ABSENT"
	providerReference := result.ProviderRequestID
	resultReference := "documents/deletion/" + deletionID.String() + "/absent"
	receipt, err := effectsv1.NewSucceededReceipt(request,
		deletionReceiptID(request, code), code, &providerReference, resultReference,
		hex.EncodeToString(result.AbsenceSHA256[:]),
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func (provider *deletionProvider) failed(request effectsv1.Request,
	deletionID foundation.ID, code string, class effectsv1.ErrorClass,
) effectexecution.Result {
	providerReference := deletionID.String()
	receipt, err := effectsv1.NewFailedReceipt(request,
		deletionReceiptID(request, code), code, class, &providerReference,
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func deletionReceiptID(request effectsv1.Request, code string) string {
	return uuid.NewSHA1(deletionReceiptNamespace,
		[]byte(request.EffectID+"\x00"+code)).String()
}
