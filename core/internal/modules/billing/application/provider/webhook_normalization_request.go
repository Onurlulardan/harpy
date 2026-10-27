package providerapp

type WebhookNormalizationRequest struct {
	verified VerifiedWebhook
	body     []byte
}

func newWebhookNormalizationRequest(
	verified VerifiedWebhook,
	body []byte,
) WebhookNormalizationRequest {
	return WebhookNormalizationRequest{
		verified: verified, body: append([]byte(nil), body...),
	}
}

func (request WebhookNormalizationRequest) Verified() VerifiedWebhook {
	return request.verified
}

func (request WebhookNormalizationRequest) Body() []byte {
	return append([]byte(nil), request.body...)
}
