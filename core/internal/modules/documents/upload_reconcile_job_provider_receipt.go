package documents

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var uploadReconcileReceiptNamespace = uuid.MustParse("61fc8e15-d5d5-49ca-9701-df682257d055")

func (provider *uploadReconcileProvider) succeeded(request effectsv1.Request,
	parsed jobcontract.UploadReference, result uploadreconcileapp.Result,
) effectexecution.Result {
	if result.UploadID != foundation.ID(parsed.ResourceID) || result.State == "" {
		return effectexecution.Unknown()
	}
	code := "DOCUMENT_UPLOAD_RECONCILED"
	if result.Noop {
		code = "DOCUMENT_UPLOAD_RECONCILE_NOOP"
	}
	providerReference := parsed.ResourceID.String()
	resultReference := "documents/upload-reconcile/" + providerReference + "/" +
		strconv.FormatUint(parsed.Generation, 10)
	digest := sha256.Sum256([]byte(resultReference + "\x00" + string(result.State) +
		"\x00" + strconv.FormatBool(result.ObjectFound) + "\x00" +
		result.DeletionID.String()))
	receipt, err := effectsv1.NewSucceededReceipt(request,
		uploadReconcileReceiptID(request, code), code, &providerReference,
		resultReference, hex.EncodeToString(digest[:]),
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func (provider *uploadReconcileProvider) failed(request effectsv1.Request,
	uploadID foundation.ID, code string, class effectsv1.ErrorClass,
) effectexecution.Result {
	providerReference := uploadID.String()
	receipt, err := effectsv1.NewFailedReceipt(request,
		uploadReconcileReceiptID(request, code), code, class, &providerReference,
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func uploadReconcileReceiptID(request effectsv1.Request, code string) string {
	return uuid.NewSHA1(uploadReconcileReceiptNamespace,
		[]byte(request.EffectID+"\x00"+code)).String()
}
