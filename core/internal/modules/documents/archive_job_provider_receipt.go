package documents

import (
	"encoding/hex"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var archiveReceiptNamespace = uuid.MustParse("3f85b3f5-19c4-44b5-b0ad-e0b23bde532e")

func (provider *archiveProvider) succeeded(request effectsv1.Request,
	archiveID foundation.ID, snapshot archiveapp.Snapshot,
) effectexecution.Result {
	if snapshot.ID != archiveID || snapshot.State != archiveapp.StateReady ||
		snapshot.CiphertextSHA256 == [32]byte{} {
		return effectexecution.Unknown()
	}
	code := "DOCUMENT_ARCHIVE_READY"
	receiptID := archiveReceiptID(request, code)
	providerReference := archiveID.String()
	resultReference := "documents/archive/" + archiveID.String() + "/ready"
	receipt, err := effectsv1.NewSucceededReceipt(request, receiptID, code,
		&providerReference, resultReference,
		hex.EncodeToString(snapshot.CiphertextSHA256[:]),
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func (provider *archiveProvider) failed(request effectsv1.Request,
	archiveID foundation.ID, code string, class effectsv1.ErrorClass,
) effectexecution.Result {
	providerReference := archiveID.String()
	receipt, err := effectsv1.NewFailedReceipt(request,
		archiveReceiptID(request, code), code, class, &providerReference,
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func archiveReceiptID(request effectsv1.Request, code string) string {
	return uuid.NewSHA1(archiveReceiptNamespace,
		[]byte(request.EffectID+"\x00"+code)).String()
}

func archiveDefinitive(receipt effectsv1.Receipt, err error) effectexecution.Result {
	if err != nil {
		return effectexecution.Unknown()
	}
	result, err := effectexecution.Definitive(receipt)
	if err != nil {
		return effectexecution.Unknown()
	}
	return result
}
