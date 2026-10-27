package providerapp

import "testing"

func TestIngressAdvertisesOnlyConfiguredWebhookCapabilities(t *testing.T) {
	verified := verifiedWebhookFixture(t)
	registry, _, _ := webhookRegistryFixture(
		t, verified, normalizedWebhookFixture(t, verified),
	)
	ingress, err := NewIngress(
		registry, webhookClock{webhookNow}, webhookIDs{webhookReceiptID},
		&webhookProtector{}, &ingressFixtureStore{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !ingress.SupportsWebhook("testpay") ||
		ingress.SupportsWebhook("manual") ||
		ingress.SupportsWebhook("unknown") ||
		ingress.SupportsWebhook("Testpay") {
		t.Fatal("webhook capability preflight is not fail-closed")
	}
}
