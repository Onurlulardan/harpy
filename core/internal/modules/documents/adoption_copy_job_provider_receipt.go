package documents

import (
	"encoding/hex"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var adoptionCopyReceiptNamespace = uuid.MustParse("943127cb-5b0c-4e15-b14b-c3f4af01018d")

func (provider *adoptionCopyProvider) succeeded(request effectsv1.Request,
	parsed jobcontract.AdoptionCopyReference, checkpoint adoptionapp.Checkpoint,
) effectexecution.Result {
	if checkpoint.RunID != foundation.ID(parsed.ResourceID) ||
		checkpoint.SourceVersionID != foundation.ID(parsed.SourceVersionID) ||
		checkpoint.State != adoptionapp.CheckpointVerified ||
		checkpoint.TargetObject.Version == "" ||
		checkpoint.TargetObject.CiphertextSHA256 == [32]byte{} {
		return effectexecution.Unknown()
	}
	code := "DOCUMENT_ADOPTION_COPY_VERIFIED"
	providerReference := checkpoint.TargetObject.Version
	resultReference := "documents/adoption/" + parsed.ResourceID.String() +
		"/copies/" + parsed.SourceVersionID.String() + "/verified"
	receipt, err := effectsv1.NewSucceededReceipt(request,
		adoptionCopyReceiptID(request, code), code, &providerReference,
		resultReference, hex.EncodeToString(checkpoint.TargetObject.CiphertextSHA256[:]),
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func (provider *adoptionCopyProvider) failed(request effectsv1.Request,
	parsed jobcontract.AdoptionCopyReference, code string,
	class effectsv1.ErrorClass,
) effectexecution.Result {
	providerReference := parsed.SourceVersionID.String()
	receipt, err := effectsv1.NewFailedReceipt(request,
		adoptionCopyReceiptID(request, code), code, class, &providerReference,
		provider.now().UTC().Format(time.RFC3339Nano))
	return archiveDefinitive(receipt, err)
}

func adoptionCopyReceiptID(request effectsv1.Request, code string) string {
	return uuid.NewSHA1(adoptionCopyReceiptNamespace,
		[]byte(request.EffectID+"\x00"+code)).String()
}
