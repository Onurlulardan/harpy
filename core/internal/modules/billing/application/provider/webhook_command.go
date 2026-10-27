package providerapp

import domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"

type ReceiveCommand struct {
	Provider domainprovider.Key
	Headers  Headers
	Body     []byte
}

func NewWebhookCommand(
	key domainprovider.Key,
	headers Headers,
	body []byte,
) (ReceiveCommand, error) {
	if !key.Valid() || len(body) == 0 || len(body) > maximumWebhookBodyBytes ||
		len(headers.values) == 0 {
		return ReceiveCommand{}, ErrWebhookRequest
	}
	return ReceiveCommand{
		Provider: key, Headers: cloneHeaders(headers), Body: append([]byte(nil), body...),
	}, nil
}

func (command ReceiveCommand) valid() bool {
	validated, err := NewWebhookCommand(command.Provider, command.Headers, command.Body)
	return err == nil && validated.Provider == command.Provider &&
		string(validated.Body) == string(command.Body)
}
