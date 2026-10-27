package providerapp

func (ingress *Ingress) SupportsWebhook(provider string) bool {
	if ingress == nil {
		return false
	}
	name, err := CanonicalProviderName(provider)
	if err != nil || name != provider {
		return false
	}
	for _, candidate := range ingress.registry.WebhookProviders() {
		if candidate == provider {
			return true
		}
	}
	return false
}
