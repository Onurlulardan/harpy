package application

import (
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (worker *DeliveryWorker) normalizeResult(result ports.ProviderResult,
	startedAt time.Time,
) ports.ProviderResult {
	result.Code = strings.TrimSpace(result.Code)
	result.ResponseClass = strings.TrimSpace(result.ResponseClass)
	result.StatusCode = strings.TrimSpace(result.StatusCode)
	invalid := result.Code == "" || len(result.ResponseClass) > 64 ||
		len(result.StatusCode) > 32 || result.CompletedAt.IsZero() ||
		result.CompletedAt.Before(startedAt) ||
		result.CompletedAt.After(worker.clock.Now().UTC().Add(time.Minute))
	if result.Outcome == ports.ProviderAccepted &&
		result.ProviderReferenceDigest == [32]byte{} {
		invalid = true
	}
	if invalid || (result.Outcome != ports.ProviderAccepted &&
		result.Outcome != ports.ProviderTransient &&
		result.Outcome != ports.ProviderPermanent &&
		result.Outcome != ports.ProviderUnknown) {
		return ports.ProviderResult{Outcome: ports.ProviderUnknown,
			Code: "NOTIFICATION_OUTCOME_UNKNOWN", ResponseClass: "invalid_result",
			CompletedAt: worker.clock.Now().UTC()}
	}
	return result
}
