package providerapp

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestReceiveWebhookDurablyAcceptsMappedTenantWithoutTenantInput(t *testing.T) {
	verified := verifiedWebhookFixture(t)
	registry, key, _ := webhookRegistryFixture(
		t, verified, normalizedWebhookFixture(t, verified),
	)
	protector := &webhookProtector{}
	store := &ingressFixtureStore{decision: WebhookAcceptDecision{
		ReceiptID: webhookReceiptID, TenantID: webhookTenantID, Created: true,
	}}
	ingress, err := NewIngress(
		registry, webhookClock{webhookNow}, webhookIDs{webhookReceiptID},
		protector, store,
	)
	if err != nil {
		t.Fatal(err)
	}
	headers, _ := NewHeaders(map[string][]string{
		"Testpay-Signature": {"signed"},
		"X-Tenant-ID":       {uuid.NewString()},
	})
	command, _ := NewWebhookCommand(key, headers, []byte(`{"id":"evt_100"}`))
	result, err := ingress.ReceiveWebhook(context.Background(), command)
	if err != nil || result.ReceiptID != webhookReceiptID || !result.Created {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	accepted := store.acceptance
	if store.calls != 1 || protector.called != 1 ||
		accepted.Provider() != key || accepted.Verified() != verified ||
		accepted.ProcessorJob().Policy() != WebhookProcessorPolicy() ||
		string(accepted.Body().Ciphertext()) == string(command.Body) {
		t.Fatalf("invalid durable acceptance: %+v", accepted)
	}
}

func TestReceiveWebhookTreatsSameHashDuplicateAsAccepted(t *testing.T) {
	verified := verifiedWebhookFixture(t)
	registry, key, _ := webhookRegistryFixture(
		t, verified, normalizedWebhookFixture(t, verified),
	)
	existingReceipt := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	store := &ingressFixtureStore{decision: WebhookAcceptDecision{
		ReceiptID: existingReceipt, TenantID: webhookTenantID, Created: false,
	}}
	ingress, _ := NewIngress(
		registry, webhookClock{webhookNow}, webhookIDs{webhookReceiptID},
		&webhookProtector{}, store,
	)
	headers, _ := NewHeaders(map[string][]string{"Testpay-Signature": {"signed"}})
	command, _ := NewWebhookCommand(key, headers, []byte(`{"id":"evt_100"}`))
	result, err := ingress.ReceiveWebhook(context.Background(), command)
	if err != nil || result.ReceiptID != existingReceipt || result.Created {
		t.Fatalf("duplicate result=%+v error=%v", result, err)
	}
	store.err = ErrWebhookConflict
	if _, err := ingress.ReceiveWebhook(context.Background(), command); !errors.Is(
		err, ErrWebhookConflict,
	) {
		t.Fatalf("reuse conflict error=%v", err)
	}
}

func TestReceiveWebhookFailsBeforeDurableAcceptance(t *testing.T) {
	verified := verifiedWebhookFixture(t)
	registry, key, adapter := webhookRegistryFixture(
		t, verified, normalizedWebhookFixture(t, verified),
	)
	protector := &webhookProtector{}
	store := &ingressFixtureStore{}
	ingress, _ := NewIngress(
		registry, webhookClock{webhookNow}, webhookIDs{webhookReceiptID},
		protector, store,
	)
	headers, _ := NewHeaders(map[string][]string{"Testpay-Signature": {"bad"}})
	command, _ := NewWebhookCommand(key, headers, []byte(`{"id":"evt_100"}`))
	adapter.verification = ErrWebhookAuthentication
	if _, err := ingress.ReceiveWebhook(context.Background(), command); !errors.Is(
		err, ErrWebhookAuthentication,
	) {
		t.Fatalf("authentication error=%v", err)
	}
	if protector.called != 0 || store.calls != 0 {
		t.Fatal("unverified webhook reached encryption or durable acceptance")
	}
	if _, err := NewWebhookCommand(
		key, headers, make([]byte, maximumWebhookBodyBytes+1),
	); !errors.Is(err, ErrWebhookRequest) {
		t.Fatalf("oversized body error=%v", err)
	}
}

func TestReceiveWebhookRejectsProtectorDigestDrift(t *testing.T) {
	verified := verifiedWebhookFixture(t)
	registry, key, _ := webhookRegistryFixture(
		t, verified, normalizedWebhookFixture(t, verified),
	)
	protector := &webhookProtector{digestOffset: 1}
	store := &ingressFixtureStore{}
	ingress, _ := NewIngress(
		registry, webhookClock{webhookNow}, webhookIDs{webhookReceiptID},
		protector, store,
	)
	headers, _ := NewHeaders(map[string][]string{"Testpay-Signature": {"signed"}})
	command, _ := NewWebhookCommand(key, headers, []byte(`{"id":"evt_100"}`))
	if _, err := ingress.ReceiveWebhook(context.Background(), command); !errors.Is(
		err, ErrProviderDependency,
	) {
		t.Fatalf("digest drift error=%v", err)
	}
	if store.calls != 0 {
		t.Fatal("invalid envelope reached durable acceptance")
	}
}
