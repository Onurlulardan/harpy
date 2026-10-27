package providerapp

import (
	"context"
	"errors"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestExecuteWebhookAppliesNormalizedEventOrQueuesReconciliation(t *testing.T) {
	body := []byte(`{"id":"evt_100","data":{"object":"sub_20"}}`)
	verified := verifiedWebhookFixture(t)
	event := normalizedWebhookFixture(t, verified)
	registry, key, _ := webhookRegistryFixture(t, verified, event)
	stored, err := NewStoredWebhook(
		webhookReceiptID, webhookTenantID, key, verified,
		protectedWebhookFixture(t, body), webhookNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	store := &processorFixtureStore{
		stored:   stored,
		decision: WebhookProcessDecision{State: WebhookReconcileQueued},
	}
	processor, _ := NewProcessor(
		registry, webhookClock{webhookNow}, &bodyFixtureOpener{body: body}, store,
	)
	decision, err := processor.ExecuteWebhook(context.Background(), webhookReceiptID)
	if err != nil || decision.State != WebhookReconcileQueued {
		t.Fatalf("decision=%+v error=%v", decision, err)
	}
	application := store.application
	job := application.ReconciliationJob()
	if store.applyCalls != 1 || application.TenantID() != webhookTenantID ||
		application.Provider() != key ||
		string(application.Event().CanonicalFact()) != `{"state":"active"}` ||
		job.Policy() != ProviderReconciliationPolicy() ||
		job.Policy().EffectClass != jobsv1.EffectExternal || job.EffectKey() == "" {
		t.Fatalf("invalid application or reconciliation job: %+v", application)
	}
}

func TestExecuteWebhookRejectsUnverifiedTimeAndInvalidDecision(t *testing.T) {
	body := []byte(`{"id":"evt_100"}`)
	verified := verifiedWebhookFixture(t)
	wrongTimeEvent, _ := NewNormalizedEvent(
		"billing.subscription.snapshot", 1, verified.OccurredAt().Add(-time.Second),
		ReconciliationOrdering(), []byte(`{"state":"active"}`),
	)
	registry, key, _ := webhookRegistryFixture(t, verified, wrongTimeEvent)
	stored, _ := NewStoredWebhook(
		webhookReceiptID, webhookTenantID, key, verified,
		protectedWebhookFixture(t, body), webhookNow,
	)
	store := &processorFixtureStore{stored: stored}
	processor, _ := NewProcessor(
		registry, webhookClock{webhookNow}, &bodyFixtureOpener{body: body}, store,
	)
	if _, err := processor.ExecuteWebhook(context.Background(), webhookReceiptID); !errors.Is(
		err, ErrNormalizedEvent,
	) {
		t.Fatalf("unverified normalized time error=%v", err)
	}
	if store.applyCalls != 0 {
		t.Fatal("invalid normalized event reached durable apply")
	}
	validEvent := normalizedWebhookFixture(t, verified)
	registry, _, _ = webhookRegistryFixture(t, verified, validEvent)
	store.decision = WebhookProcessDecision{}
	processor, _ = NewProcessor(
		registry, webhookClock{webhookNow}, &bodyFixtureOpener{body: body}, store,
	)
	if _, err := processor.ExecuteWebhook(context.Background(), webhookReceiptID); !errors.Is(
		err, ErrWebhookDecision,
	) {
		t.Fatalf("invalid durable decision error=%v", err)
	}
}

func TestExecuteWebhookDoesNotApplyUnreadableEnvelope(t *testing.T) {
	body := []byte(`{"id":"evt_100"}`)
	verified := verifiedWebhookFixture(t)
	event := normalizedWebhookFixture(t, verified)
	registry, key, _ := webhookRegistryFixture(t, verified, event)
	stored, _ := NewStoredWebhook(
		webhookReceiptID, webhookTenantID, key, verified,
		protectedWebhookFixture(t, body), webhookNow,
	)
	store := &processorFixtureStore{stored: stored}
	processor, _ := NewProcessor(
		registry, webhookClock{webhookNow},
		&bodyFixtureOpener{err: errors.New("key unavailable")}, store,
	)
	if _, err := processor.ExecuteWebhook(context.Background(), webhookReceiptID); !errors.Is(
		err, ErrProviderDependency,
	) {
		t.Fatalf("unreadable envelope error=%v", err)
	}
	if store.applyCalls != 0 {
		t.Fatal("unreadable body reached durable apply")
	}
}

func TestExecuteWebhookCannotApplyUnorderedProviderFact(t *testing.T) {
	body := []byte(`{"id":"evt_100"}`)
	verified := verifiedWebhookFixture(t)
	event, _ := NewNormalizedEvent(
		"billing.subscription.snapshot", 1, verified.OccurredAt(),
		ReconciliationOrdering(), []byte(`{"state":"active"}`),
	)
	registry, key, _ := webhookRegistryFixture(t, verified, event)
	stored, _ := NewStoredWebhook(
		webhookReceiptID, webhookTenantID, key, verified,
		protectedWebhookFixture(t, body), webhookNow,
	)
	store := &processorFixtureStore{
		stored: stored, decision: WebhookProcessDecision{State: WebhookApplied},
	}
	processor, _ := NewProcessor(
		registry, webhookClock{webhookNow}, &bodyFixtureOpener{body: body}, store,
	)
	if _, err := processor.ExecuteWebhook(context.Background(), webhookReceiptID); !errors.Is(
		err, ErrWebhookDecision,
	) {
		t.Fatalf("unordered apply error=%v", err)
	}
}

func TestExecuteWebhookCannotPersistRawProviderBodyAsNormalizedFact(t *testing.T) {
	body := []byte(`{"id":"evt_100"}`)
	verified := verifiedWebhookFixture(t)
	ordering, _ := ContiguousOrdering(1, 0)
	event, _ := NewNormalizedEvent(
		"billing.subscription.snapshot", 1, verified.OccurredAt(), ordering, body,
	)
	registry, key, _ := webhookRegistryFixture(t, verified, event)
	stored, _ := NewStoredWebhook(
		webhookReceiptID, webhookTenantID, key, verified,
		protectedWebhookFixture(t, body), webhookNow,
	)
	store := &processorFixtureStore{stored: stored}
	processor, _ := NewProcessor(
		registry, webhookClock{webhookNow}, &bodyFixtureOpener{body: body}, store,
	)
	if _, err := processor.ExecuteWebhook(context.Background(), webhookReceiptID); !errors.Is(
		err, ErrNormalizedEvent,
	) {
		t.Fatalf("raw normalized body error=%v", err)
	}
	if store.applyCalls != 0 {
		t.Fatal("raw provider body reached normalized event storage")
	}
}
