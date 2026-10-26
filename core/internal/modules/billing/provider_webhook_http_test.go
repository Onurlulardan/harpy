package billing

import (
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func TestProviderWebhookHTTPErrorVocabularyIsStable(t *testing.T) {
	tests := []struct {
		provider error
		http     error
	}{
		{providerapp.ErrCapability, billingapi.ErrNotFound},
		{providerapp.ErrWebhookAuthentication, billingapi.ErrAuthentication},
		{providerapp.ErrWebhookRequest, billingapi.ErrInvalid},
		{providerapp.ErrWebhookMalformed, billingapi.ErrInvalid},
		{providerapp.ErrWebhookMapping, billingapi.ErrNotFound},
		{providerapp.ErrWebhookReplay, billingapi.ErrConflict},
		{providerapp.ErrWebhookConflict, billingapi.ErrConflict},
		{providerapp.ErrProviderDependency, billingapi.ErrUnavailable},
	}
	for _, test := range tests {
		if got := webhookHTTPError(test.provider); !errors.Is(got, test.http) {
			t.Fatalf("provider error %v mapped to %v, want %v", test.provider, got, test.http)
		}
	}
}

func TestProviderWebhookHTTPIngressRejectsMissingApplication(t *testing.T) {
	if _, err := NewProviderWebhookHTTPIngress(nil); !errors.Is(
		err, providerapp.ErrConfiguration,
	) {
		t.Fatalf("missing provider ingress error = %v", err)
	}
}
