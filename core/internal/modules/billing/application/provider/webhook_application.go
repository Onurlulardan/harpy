package providerapp

import (
	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

type WebhookApplication struct {
	receiptID      uuid.UUID
	tenantID       uuid.UUID
	provider       domainprovider.Key
	verified       VerifiedWebhook
	event          NormalizedEvent
	reconciliation AsyncJob
}

func newWebhookApplication(
	stored StoredWebhook,
	event NormalizedEvent,
	reconciliation AsyncJob,
) (WebhookApplication, error) {
	if !stored.valid() || !event.valid() || !reconciliation.valid() ||
		reconciliation.Policy() != ProviderReconciliationPolicy() {
		return WebhookApplication{}, ErrWebhookProcessing
	}
	return WebhookApplication{
		receiptID: stored.receiptID, tenantID: stored.tenantID,
		provider: stored.provider, verified: stored.verified,
		event: cloneNormalizedEvent(event), reconciliation: reconciliation,
	}, nil
}

func (value WebhookApplication) ReceiptID() uuid.UUID         { return value.receiptID }
func (value WebhookApplication) TenantID() uuid.UUID          { return value.tenantID }
func (value WebhookApplication) Provider() domainprovider.Key { return value.provider }
func (value WebhookApplication) Verified() VerifiedWebhook    { return value.verified }
func (value WebhookApplication) Event() NormalizedEvent       { return cloneNormalizedEvent(value.event) }
func (value WebhookApplication) ReconciliationJob() AsyncJob  { return value.reconciliation }

func cloneNormalizedEvent(value NormalizedEvent) NormalizedEvent {
	copy, _ := NewNormalizedEvent(
		value.schema, value.revision, value.occurredAt, value.ordering, value.fact,
	)
	return copy
}
