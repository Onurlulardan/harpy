package notifications

import (
	"context"
	"errors"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (provider *deliveryProvider) Execute(ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	reference, err := provider.validateRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	processed, err := provider.worker.Execute(ctx, reference.tenantID,
		reference.deliveryID, request.EffectID)
	if errors.Is(err, application.ErrOutcomeUnknown) {
		return effectexecution.Unknown()
	}
	if err != nil {
		return provider.failed(request, reference, "DELIVERY_STATE_UNAVAILABLE",
			effectsv1.ErrorTransient)
	}
	switch processed.Outcome {
	case ports.ProviderAccepted:
		return provider.succeeded(request, reference, processed.Code)
	case ports.ProviderTransient:
		return provider.failed(request, reference, processed.Code, effectsv1.ErrorTransient)
	case ports.ProviderPermanent:
		if processed.Code == "PREFERENCE_DISABLED" {
			return provider.succeeded(request, reference, "DELIVERY_SUPPRESSED")
		}
		return provider.failed(request, reference, processed.Code, effectsv1.ErrorPermanent)
	default:
		return effectexecution.Unknown()
	}
}
