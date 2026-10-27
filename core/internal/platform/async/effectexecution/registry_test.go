package effectexecution_test

import (
	"context"
	"errors"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution/testadapter"
)

const (
	tenantID  = "11111111-1111-4111-8111-111111111111"
	effectID  = "22222222-2222-4222-8222-222222222222"
	causation = "33333333-3333-4333-8333-333333333333"
	receiptID = "44444444-4444-4444-8444-444444444444"
	digest    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestRegistryInvokesExactProviderWithDefensiveRequest(t *testing.T) {
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	adapter, _ := testadapter.New(identity, func(request effectsv1.Request) effectexecution.Result {
		receipt, err := effectsv1.NewSucceededReceipt(
			request, receiptID, "MAIL_SENT", nil, "result://mail/1", digest,
			"2026-07-15T10:00:01Z",
		)
		if err != nil {
			t.Fatal(err)
		}
		result, err := effectexecution.Definitive(receipt)
		if err != nil {
			t.Fatal(err)
		}
		return result
	})
	registry, err := effectexecution.NewRegistry([]effectexecution.Provider{adapter})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := registry.Resolve(identity)
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolved.Invoke(context.Background(), requestFixture(t))
	if err != nil || result.IsUnknown() {
		t.Fatalf("result=%v error=%v", result, err)
	}
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "MAIL_SENT" || len(adapter.Requests()) != 1 {
		t.Fatal("definitive provider result was not retained")
	}
}

func TestRegistryFailsClosedForMissingOrMalformedProvider(t *testing.T) {
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	registry, _ := effectexecution.NewRegistry(nil)
	if _, err := registry.Resolve(identity); !errors.Is(err, effectexecution.ErrProviderNotRegistered) {
		t.Fatalf("missing provider error = %v", err)
	}
	adapter, _ := testadapter.New(identity, func(effectsv1.Request) effectexecution.Result {
		return effectexecution.Result{}
	})
	registry, _ = effectexecution.NewRegistry([]effectexecution.Provider{adapter})
	resolved, _ := registry.Resolve(identity)
	if _, err := resolved.Invoke(context.Background(), requestFixture(t)); !errors.Is(err, effectexecution.ErrProviderResult) {
		t.Fatalf("malformed result error = %v", err)
	}
}

func requestFixture(t *testing.T) effectsv1.Request {
	t.Helper()
	request, err := effectsv1.NewRequest(
		effectID, "notification.email", 1, stringPointer(tenantID), effectsv1.Authority{
			ScopeKind: effectsv1.ScopeTenant, ScopeID: tenantID,
			GroupID: "tenant_core", GroupGeneration: 1,
			Domain: "notifications", CutoverGeneration: 1,
			WriterOrigin: effectsv1.WriterV2,
		}, "test.mail", "send", "mail:1", "corr-1", causation,
		"payload://mail/1", digest, "2026-07-15T10:00:00Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func stringPointer(value string) *string { return &value }
