package providerapp

import domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"

func CanonicalProviderName(value string) (string, error) {
	key, err := domainprovider.NewKey(value)
	if err != nil {
		return "", ErrCapability
	}
	return string(key), nil
}
