package providerapp

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
)

type manualFixtureAdapter struct{}

func (*manualFixtureAdapter) Descriptor() domainprovider.Descriptor {
	return domainprovider.ManualDescriptor()
}

func (*manualFixtureAdapter) ExecuteEffect(
	_ context.Context,
	request effectsv1.Request,
) (EffectResult, error) {
	return UnknownEffect(request)
}

func (*manualFixtureAdapter) Reconcile(
	_ context.Context,
	request effectsv1.Request,
) (EffectResult, error) {
	return UnknownEffect(request)
}

type externalFixtureAdapter struct {
	descriptor    domainprovider.Descriptor
	verified      VerifiedWebhook
	verification  error
	event         NormalizedEvent
	normalization error
	verifyCalls   atomic.Int64
}

func (adapter *externalFixtureAdapter) Descriptor() domainprovider.Descriptor {
	return adapter.descriptor
}

func (adapter *externalFixtureAdapter) VerifyWebhook(
	_ context.Context,
	_ WebhookVerificationRequest,
) (VerifiedWebhook, error) {
	adapter.verifyCalls.Add(1)
	return adapter.verified, adapter.verification
}

func (adapter *externalFixtureAdapter) NormalizeWebhook(
	_ WebhookNormalizationRequest,
) (NormalizedEvent, error) {
	return adapter.event, adapter.normalization
}

func (adapter *externalFixtureAdapter) ExecuteEffect(
	_ context.Context,
	request effectsv1.Request,
) (EffectResult, error) {
	return UnknownEffect(request)
}

func (adapter *externalFixtureAdapter) Reconcile(
	_ context.Context,
	request effectsv1.Request,
) (EffectResult, error) {
	return UnknownEffect(request)
}

func externalFixture(t *testing.T) (domainprovider.Key, *externalFixtureAdapter) {
	t.Helper()
	key, err := domainprovider.NewKey("testpay")
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := domainprovider.NewDescriptor(key, domainprovider.Capabilities{
		Commands: true, Webhooks: true, Reconciliation: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return key, &externalFixtureAdapter{descriptor: descriptor}
}

func registryFixture(t *testing.T) (Registry, domainprovider.Key, *externalFixtureAdapter) {
	t.Helper()
	key, external := externalFixture(t)
	registry, err := NewRegistry(&manualFixtureAdapter{}, external)
	if err != nil {
		t.Fatal(err)
	}
	return registry, key, external
}

type capabilityMismatchAdapter struct{ *externalFixtureAdapter }

func (adapter *capabilityMismatchAdapter) Descriptor() domainprovider.Descriptor {
	descriptor, _ := domainprovider.NewDescriptor(
		adapter.externalFixtureAdapter.descriptor.Key(),
		domainprovider.Capabilities{Commands: true},
	)
	return descriptor
}

func assertErrorIs(t *testing.T, err error, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("error=%v, want %v", err, target)
	}
}
