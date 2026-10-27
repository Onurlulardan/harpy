package providerapp

import (
	"context"

	"github.com/google/uuid"
)

type WebhookProcessorStore interface {
	LoadWebhook(context.Context, uuid.UUID) (StoredWebhook, error)
	ApplyWebhook(context.Context, WebhookApplication) (WebhookProcessDecision, error)
}
