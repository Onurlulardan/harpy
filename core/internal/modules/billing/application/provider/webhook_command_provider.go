package providerapp

import domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"

func NewWebhookCommandForProvider(
	provider string,
	headers Headers,
	body []byte,
) (ReceiveCommand, error) {
	key, err := domainprovider.NewKey(provider)
	if err != nil {
		return ReceiveCommand{}, ErrCapability
	}
	return NewWebhookCommand(key, headers, body)
}
