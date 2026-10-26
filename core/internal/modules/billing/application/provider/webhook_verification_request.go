package providerapp

import "time"

type WebhookVerificationRequest struct {
	headers    Headers
	body       []byte
	receivedAt time.Time
}

func newWebhookVerificationRequest(
	headers Headers,
	body []byte,
	receivedAt time.Time,
) WebhookVerificationRequest {
	return WebhookVerificationRequest{
		headers: cloneHeaders(headers), body: append([]byte(nil), body...),
		receivedAt: receivedAt,
	}
}

func (request WebhookVerificationRequest) Headers() Headers {
	return cloneHeaders(request.headers)
}

func (request WebhookVerificationRequest) Body() []byte {
	return append([]byte(nil), request.body...)
}

func (request WebhookVerificationRequest) ReceivedAt() time.Time {
	return request.receivedAt
}
