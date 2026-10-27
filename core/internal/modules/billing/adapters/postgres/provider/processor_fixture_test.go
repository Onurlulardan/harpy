package provider

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

type fixtureClock struct{ now time.Time }

func (clock fixtureClock) Now() time.Time { return clock.now }

type fixtureOpener struct{ body []byte }

func (opener fixtureOpener) OpenForReconciliation(
	_ context.Context,
	_ []byte,
	_ providerapp.ProtectedBody,
) ([]byte, error) {
	return append([]byte(nil), opener.body...), nil
}

type fixtureWebhookAdapter struct {
	descriptor domainprovider.Descriptor
	event      providerapp.NormalizedEvent
}

type fixtureManualAdapter struct{}

func (fixtureManualAdapter) Descriptor() domainprovider.Descriptor {
	return domainprovider.ManualDescriptor()
}

func (fixtureManualAdapter) ExecuteEffect(
	context.Context,
	effectsv1.Request,
) (providerapp.EffectResult, error) {
	return providerapp.EffectResult{}, errors.New("manual effect is not used")
}

func (fixtureManualAdapter) Reconcile(
	context.Context,
	effectsv1.Request,
) (providerapp.EffectResult, error) {
	return providerapp.EffectResult{}, errors.New("manual reconciliation is not used")
}

func (adapter *fixtureWebhookAdapter) Descriptor() domainprovider.Descriptor {
	return adapter.descriptor
}

func (adapter *fixtureWebhookAdapter) VerifyWebhook(
	context.Context,
	providerapp.WebhookVerificationRequest,
) (providerapp.VerifiedWebhook, error) {
	return providerapp.VerifiedWebhook{}, errors.New("verification is not used")
}

func (adapter *fixtureWebhookAdapter) NormalizeWebhook(
	providerapp.WebhookNormalizationRequest,
) (providerapp.NormalizedEvent, error) {
	return adapter.event, nil
}

func fixtureRegistry(
	t *testing.T,
	event providerapp.NormalizedEvent,
) (providerapp.Registry, domainprovider.Key) {
	t.Helper()
	key, _ := domainprovider.NewKey("testpay")
	descriptor, _ := domainprovider.NewDescriptor(
		key, domainprovider.Capabilities{Webhooks: true},
	)
	registry, err := providerapp.NewRegistry(
		fixtureManualAdapter{}, &fixtureWebhookAdapter{descriptor: descriptor, event: event},
	)
	if err != nil {
		t.Fatal(err)
	}
	return registry, key
}

func fixtureWebhookRows(
	authorityID, mappingID, accountID, subscriptionID uuid.UUID,
	now time.Time,
	raw []byte,
	state string,
) []any {
	authentication := sha256.Sum256([]byte("authentication"))
	bodySHA := sha256.Sum256(raw)
	return []any{
		authorityID, mappingID, accountID, "testpay", "acct_10",
		"subscription", "sub_20", &subscriptionID, (*uuid.UUID)(nil), "evt_30",
		authentication[:], 1, nullableTime{value: now.Add(-time.Second), valid: true},
		nullableTime{value: now.Add(-2 * time.Second), valid: true},
		nullableString{value: "AES-256-GCM", valid: true}, []byte("123456789012"),
		make([]byte, 17), []byte("wrapped"),
		nullableString{value: "billing-kek-v1", valid: true}, bodySHA[:],
		nullableTime{value: now, valid: true}, nullableString{value: state, valid: true},
	}
}
