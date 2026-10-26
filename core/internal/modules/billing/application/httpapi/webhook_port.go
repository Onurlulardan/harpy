package httpapi

import "context"

type WebhookIngress interface {
	AcceptsProvider(string) bool
	ReceiveWebhook(context.Context, WebhookRequest) (WebhookAcceptance, error)
}
