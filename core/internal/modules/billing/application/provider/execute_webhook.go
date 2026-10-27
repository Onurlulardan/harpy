package providerapp

import (
	"context"
	"crypto/sha256"

	"github.com/google/uuid"
)

type Processor struct {
	registry Registry
	clock    Clock
	opener   ReconciliationBodyOpener
	store    WebhookProcessorStore
}

func NewProcessor(
	registry Registry,
	clock Clock,
	opener ReconciliationBodyOpener,
	store WebhookProcessorStore,
) (*Processor, error) {
	if registry.adapters == nil || nilDependency(clock) || nilDependency(opener) ||
		nilDependency(store) {
		return nil, ErrConfiguration
	}
	return &Processor{registry: registry, clock: clock, opener: opener, store: store}, nil
}

func (processor *Processor) ExecuteWebhook(
	ctx context.Context,
	receiptID uuid.UUID,
) (WebhookProcessDecision, error) {
	if processor == nil || ctx == nil || ctx.Err() != nil || receiptID == uuid.Nil {
		return WebhookProcessDecision{}, ErrWebhookProcessing
	}
	stored, err := processor.store.LoadWebhook(ctx, receiptID)
	if err != nil {
		return WebhookProcessDecision{}, err
	}
	if !stored.valid() || stored.ReceiptID() != receiptID {
		return WebhookProcessDecision{}, ErrWebhookProcessing
	}
	adapter, err := processor.registry.Webhook(stored.Provider())
	if err != nil {
		return WebhookProcessDecision{}, err
	}
	associatedData := webhookAssociatedData(receiptID, stored.Provider(), stored.Verified())
	body, err := processor.opener.OpenForReconciliation(ctx, associatedData, stored.Body())
	if err != nil {
		return WebhookProcessDecision{}, ErrProviderDependency
	}
	defer clear(body)
	if len(body) == 0 || len(body) > maximumWebhookBodyBytes ||
		sha256.Sum256(body) != stored.Body().BodySHA256() {
		return WebhookProcessDecision{}, ErrWebhookProcessing
	}
	event, err := adapter.NormalizeWebhook(
		newWebhookNormalizationRequest(stored.Verified(), body),
	)
	if err != nil {
		return WebhookProcessDecision{}, err
	}
	if !event.valid() || !event.OccurredAt().Equal(stored.Verified().OccurredAt()) ||
		event.FactSHA256() == sha256.Sum256(body) {
		return WebhookProcessDecision{}, ErrNormalizedEvent
	}
	runAt, err := canonicalInstant(processor.clock.Now())
	if err != nil {
		return WebhookProcessDecision{}, ErrProviderDependency
	}
	reconciliation, err := newProviderReconciliationJob(
		receiptID, stored.Provider(), stored.Verified(), event.FactSHA256(), runAt,
	)
	if err != nil {
		return WebhookProcessDecision{}, err
	}
	application, err := newWebhookApplication(stored, event, reconciliation)
	if err != nil {
		return WebhookProcessDecision{}, err
	}
	decision, err := processor.store.ApplyWebhook(ctx, application)
	if err != nil {
		return WebhookProcessDecision{}, err
	}
	if !decision.validFor(event.Ordering()) {
		return WebhookProcessDecision{}, ErrWebhookDecision
	}
	return decision, nil
}
