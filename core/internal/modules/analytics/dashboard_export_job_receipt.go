package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

var dashboardExportReceiptNamespace = uuid.MustParse("8e7a6d02-df6d-4327-b179-160296b50b8e")

func (provider *dashboardExportJobProvider) succeeded(request effectsv1.Request,
	exportID uuid.UUID, state exportapp.State,
) effectexecution.Result {
	code := "ANALYTICS_EXPORT_READY"
	if provider.operation == "delete" {
		code = "ANALYTICS_EXPORT_DELETED"
	} else if provider.operation == "expire" {
		code = "ANALYTICS_EXPORT_EXPIRY_COMPLETE"
	} else if state != exportapp.StateReady {
		code = "ANALYTICS_EXPORT_RENDER_TERMINAL"
	}
	providerReference := exportID.String()
	resultReference := "analytics-dashboard-export://" + provider.operation + "/" +
		exportID.String() + "/" + string(state)
	digest := sha256.Sum256([]byte(resultReference))
	receipt, err := effectsv1.NewSucceededReceipt(request,
		dashboardExportReceiptID(request, code), code, &providerReference,
		resultReference, hex.EncodeToString(digest[:]), provider.completedAt())
	return dashboardExportDefinitive(receipt, err)
}

func (provider *dashboardExportJobProvider) failed(request effectsv1.Request,
	exportID uuid.UUID, code string, class effectsv1.ErrorClass,
) effectexecution.Result {
	providerReference := exportID.String()
	receipt, err := effectsv1.NewFailedReceipt(request,
		dashboardExportReceiptID(request, code), code, class,
		&providerReference, provider.completedAt())
	return dashboardExportDefinitive(receipt, err)
}

func (provider *dashboardExportJobProvider) completedAt() string {
	now := provider.now()
	if now.IsZero() {
		return time.Time{}.Format(time.RFC3339Nano)
	}
	return now.UTC().Format(time.RFC3339Nano)
}

func dashboardExportReceiptID(request effectsv1.Request, code string) string {
	return uuid.NewSHA1(dashboardExportReceiptNamespace,
		[]byte(request.EffectID+"\x00"+code)).String()
}

func dashboardExportDefinitive(receipt effectsv1.Receipt, err error) effectexecution.Result {
	if err != nil {
		return effectexecution.Unknown()
	}
	result, err := effectexecution.Definitive(receipt)
	if err != nil {
		return effectexecution.Unknown()
	}
	return result
}

func dashboardExportReferenceDigest(reference string) string {
	digest := sha256.Sum256([]byte(reference))
	return hex.EncodeToString(digest[:])
}
