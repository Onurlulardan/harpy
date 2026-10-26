package billing

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *Handler) WebhookRoutes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{{
		Method: http.MethodPost, Pattern: "/webhooks/billing/{provider}",
		Handler: http.HandlerFunc(handler.receiveWebhook),
	}}
}
