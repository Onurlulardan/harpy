package telegram

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func invalidResult(client *Client,
	request ports.ProviderRequest,
) ports.ProviderResult {
	at := request.Requested.UTC()
	if client != nil && client.now != nil {
		at = client.now().UTC()
	}
	if at.IsZero() {
		at = time.Unix(1, 0).UTC()
	}
	return ports.ProviderResult{Outcome: ports.ProviderPermanent,
		Code: "TELEGRAM_REQUEST_INVALID", ResponseClass: "validation", CompletedAt: at}
}
