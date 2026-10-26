package providerapp

import (
	"context"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
)

type Adapter interface {
	Descriptor() domainprovider.Descriptor
}

type WebhookAdapter interface {
	Adapter
	VerifyWebhook(context.Context, WebhookVerificationRequest) (VerifiedWebhook, error)
	NormalizeWebhook(WebhookNormalizationRequest) (NormalizedEvent, error)
}

type EffectExecutor interface {
	Adapter
	ExecuteEffect(context.Context, effectsv1.Request) (EffectResult, error)
}

type Reconciler interface {
	Adapter
	Reconcile(context.Context, effectsv1.Request) (EffectResult, error)
}
