package provider

import (
	"context"
	"fmt"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func (storage *processorStore) applySubscriptionFact(
	ctx context.Context,
	row webhookRow,
	application providerapp.WebhookApplication,
	eventID uuid.UUID,
	normalizedAt time.Time,
) (providerapp.WebhookProcessDecision, error) {
	if row.objectKind != "subscription" || row.subscriptionID == nil ||
		*row.subscriptionID == uuid.Nil {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrWebhookMapping
	}
	fact, err := decodeSubscriptionFact(application.Event().CanonicalFact())
	if err != nil || fact.subscriptionID != *row.subscriptionID {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrNormalizedEvent
	}
	if fact.boundary != nil {
		if err := storage.insertBoundaryEvidence(
			ctx, row, eventID, fact, normalizedAt,
		); err != nil {
			return providerapp.WebhookProcessDecision{}, err
		}
	}
	request := fact.request(
		application.ReceiptID(), eventID, application.Event().FactSHA256(),
		application.Event().OccurredAt(),
	)
	result, err := storage.update.UpdateFromProvider(ctx, request)
	if err != nil {
		return providerapp.WebhookProcessDecision{}, fmt.Errorf(
			"apply billing subscription provider fact: %w", err,
		)
	}
	if result.Replay || result.Subscription.ID != fact.subscriptionID ||
		result.Subscription.TenantID != storage.tenantID {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrWebhookConflict
	}
	return providerapp.WebhookProcessDecision{State: providerapp.WebhookApplied}, nil
}
