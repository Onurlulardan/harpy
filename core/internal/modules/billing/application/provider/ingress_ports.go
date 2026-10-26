package providerapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type IDSource interface {
	New() uuid.UUID
}

type WebhookIngressStore interface {
	AcceptWebhook(context.Context, WebhookAcceptance) (WebhookAcceptDecision, error)
}
