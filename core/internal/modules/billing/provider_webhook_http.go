package billing

import (
	"context"
	"errors"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

type providerWebhookHTTP struct {
	ingress *providerapp.Ingress
}

func (adapter *providerWebhookHTTP) AcceptsProvider(provider string) bool {
	return adapter != nil && adapter.ingress != nil &&
		adapter.ingress.SupportsWebhook(provider)
}

func NewProviderWebhookHTTPIngress(
	ingress *providerapp.Ingress,
) (billingapi.WebhookIngress, error) {
	if ingress == nil {
		return nil, providerapp.ErrConfiguration
	}
	return &providerWebhookHTTP{ingress: ingress}, nil
}

func (adapter *providerWebhookHTTP) ReceiveWebhook(
	ctx context.Context,
	request billingapi.WebhookRequest,
) (billingapi.WebhookAcceptance, error) {
	if adapter == nil || adapter.ingress == nil {
		return billingapi.WebhookAcceptance{}, billingapi.ErrUnavailable
	}
	headers, err := providerapp.NewHeaders(request.Headers)
	if err != nil {
		return billingapi.WebhookAcceptance{}, billingapi.ErrInvalid
	}
	command, err := providerapp.NewWebhookCommandForProvider(
		request.Provider, headers, request.Body,
	)
	if err != nil {
		return billingapi.WebhookAcceptance{}, webhookHTTPError(err)
	}
	defer clear(command.Body)
	result, err := adapter.ingress.ReceiveWebhook(ctx, command)
	if err != nil {
		return billingapi.WebhookAcceptance{}, webhookHTTPError(err)
	}
	return billingapi.WebhookAcceptance{
		Duplicate: !result.Created,
		ReceiptID: result.ReceiptID.String(),
	}, nil
}

func webhookHTTPError(err error) error {
	switch {
	case errors.Is(err, providerapp.ErrCapability):
		return billingapi.ErrNotFound
	case errors.Is(err, providerapp.ErrWebhookAuthentication):
		return billingapi.ErrAuthentication
	case errors.Is(err, providerapp.ErrWebhookRequest),
		errors.Is(err, providerapp.ErrWebhookMalformed):
		return billingapi.ErrInvalid
	case errors.Is(err, providerapp.ErrWebhookMapping):
		return billingapi.ErrNotFound
	case errors.Is(err, providerapp.ErrWebhookReplay),
		errors.Is(err, providerapp.ErrWebhookConflict):
		return billingapi.ErrConflict
	default:
		return billingapi.ErrUnavailable
	}
}
