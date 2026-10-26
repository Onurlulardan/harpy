package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type manualOnlyWebhookHTTP struct{}

func (manualOnlyWebhookHTTP) AcceptsProvider(string) bool { return false }

func NewManualOnlyWebhookHTTPIngress() billingapi.WebhookIngress {
	return manualOnlyWebhookHTTP{}
}

func (manualOnlyWebhookHTTP) ReceiveWebhook(
	context.Context,
	billingapi.WebhookRequest,
) (billingapi.WebhookAcceptance, error) {
	return billingapi.WebhookAcceptance{}, billingapi.ErrNotFound
}
