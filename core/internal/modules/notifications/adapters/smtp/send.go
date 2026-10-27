package smtp

import (
	"context"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (client *Client) Send(ctx context.Context,
	request ports.ProviderRequest,
) ports.ProviderResult {
	parsed, ok := parseRequest(request)
	if client == nil || ctx == nil || !ok {
		return ports.ProviderResult{Outcome: ports.ProviderPermanent,
			Code: "SMTP_REQUEST_INVALID", ResponseClass: "validation",
			CompletedAt: clientTime(client, request)}
	}
	defer clear(parsed.message)
	result := client.exchanger.Exchange(ctx, parsed)
	completed := result.completedAt.UTC()
	if completed.IsZero() {
		completed = client.now().UTC()
	}
	response := ports.ProviderResult{Code: result.code,
		ResponseClass: result.class, StatusCode: result.statusCode,
		CompletedAt: completed}
	if result.accepted && result.definitive {
		response.Outcome, response.Code = ports.ProviderAccepted, "SMTP_ACCEPTED"
		response.ProviderReferenceDigest = sha256.Sum256(
			[]byte(stableMessageID(request.DeliveryID.String())))
		return response
	}
	if result.definitive && result.temporary {
		response.Outcome = ports.ProviderTransient
	} else if result.definitive {
		response.Outcome = ports.ProviderPermanent
	} else if result.bytesWritten {
		response.Outcome, response.Code = ports.ProviderUnknown, "NOTIFICATION_OUTCOME_UNKNOWN"
	} else {
		response.Outcome = ports.ProviderTransient
	}
	if strings.TrimSpace(response.Code) == "" {
		response.Code = "SMTP_UNAVAILABLE"
	}
	return response
}

func clientTime(client *Client, request ports.ProviderRequest) time.Time {
	if client != nil && client.now != nil {
		return client.now().UTC()
	}
	return request.Requested.UTC()
}
