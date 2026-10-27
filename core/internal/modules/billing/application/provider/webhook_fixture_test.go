package providerapp

import (
	"context"
	"crypto/sha256"
	"time"

	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

var (
	webhookReceiptID = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	webhookTenantID  = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	webhookNow       = time.Date(2026, 7, 17, 8, 30, 0, 123000000, time.UTC)
)

type webhookClock struct{ value time.Time }

func (clock webhookClock) Now() time.Time { return clock.value }

type webhookIDs struct{ value uuid.UUID }

func (ids webhookIDs) New() uuid.UUID { return ids.value }

type webhookProtector struct {
	called       int
	digestOffset byte
}

func (protector *webhookProtector) Protect(
	_ context.Context,
	_ []byte,
	body []byte,
) (ProtectedBody, error) {
	protector.called++
	digest := sha256.Sum256(body)
	digest[0] ^= protector.digestOffset
	return NewProtectedBody(
		"AES-256-GCM", []byte("nonce"), []byte("encrypted-body"),
		"kek-v1", []byte("wrapped-key"), digest,
	)
}

type ingressFixtureStore struct {
	acceptance WebhookAcceptance
	decision   WebhookAcceptDecision
	err        error
	calls      int
}

func (store *ingressFixtureStore) AcceptWebhook(
	_ context.Context,
	acceptance WebhookAcceptance,
) (WebhookAcceptDecision, error) {
	store.calls++
	store.acceptance = acceptance
	return store.decision, store.err
}

func verifiedWebhookFixture(t testingT) VerifiedWebhook {
	t.Helper()
	authenticationSHA := sha256.Sum256([]byte("verified-signature-transcript"))
	verified, err := NewVerifiedWebhook(
		"evt_100", "acct_10", "subscription", "sub_20",
		webhookNow.Add(-time.Second), webhookNow.Add(-2*time.Second), 1,
		authenticationSHA,
	)
	if err != nil {
		t.Fatal(err)
	}
	return verified
}

func protectedWebhookFixture(t testingT, body []byte) ProtectedBody {
	t.Helper()
	digest := sha256.Sum256(body)
	protected, err := NewProtectedBody(
		"AES-256-GCM", []byte("nonce"), []byte("ciphertext"),
		"kek-v1", []byte("wrapped"), digest,
	)
	if err != nil {
		t.Fatal(err)
	}
	return protected
}

func normalizedWebhookFixture(t testingT, verified VerifiedWebhook) NormalizedEvent {
	t.Helper()
	ordering, _ := ContiguousOrdering(2, 1)
	event, err := NewNormalizedEvent(
		"billing.subscription.snapshot", 1, verified.OccurredAt(),
		ordering, []byte(`{"state":"active"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	return event
}

type testingT interface {
	Helper()
	Fatal(...any)
}

func webhookRegistryFixture(
	t testingT,
	verified VerifiedWebhook,
	event NormalizedEvent,
) (Registry, domainprovider.Key, *externalFixtureAdapter) {
	t.Helper()
	key, adapter := externalFixtureFromT(t)
	adapter.verified = verified
	adapter.event = event
	registry, err := NewRegistry(&manualFixtureAdapter{}, adapter)
	if err != nil {
		t.Fatal(err)
	}
	return registry, key, adapter
}

func externalFixtureFromT(t testingT) (domainprovider.Key, *externalFixtureAdapter) {
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
