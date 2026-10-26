package jobexecution

import (
	"context"
	"slices"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution/testadapter"
)

func TestExternalProcessorCompletesDefinitiveSuccessAfterStart(t *testing.T) {
	calls := []string{}
	policy, spec, started := externalProcessorFixture(t, &calls)
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	adapter, _ := testadapter.New(identity, func(request effectsv1.Request) effectexecution.Result {
		appendCall(&calls, "provider")
		receipt, err := effectsv1.NewSucceededReceipt(
			request, testReceiptID, "MAIL_SENT", nil,
			"result://mail/1", testDigest, "2026-07-15T10:00:01Z",
		)
		if err != nil {
			t.Fatal(err)
		}
		return definitiveResult(t, receipt)
	})
	registry, _ := effectexecution.NewRegistry([]effectexecution.Provider{adapter})
	finalizer := &externalFinalizerStub{calls: &calls}
	processor, err := NewExternalProcessor(
		policy, &externalPlannerStub{spec: spec, calls: &calls},
		&preEffectStub{started: started, calls: &calls}, finalizer,
		registry,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1); err != nil {
		t.Fatal(err)
	}
	want := []string{"plan", "start", "provider", "complete"}
	if !slices.Equal(calls, want) || finalizer.receipt.Status != effectsv1.ReceiptSucceeded {
		t.Fatalf("calls=%v receipt=%+v", calls, finalizer.receipt)
	}
}

func TestExternalProcessorFinalizesAfterProviderCancelsContext(t *testing.T) {
	calls := []string{}
	policy, spec, started := externalProcessorFixture(t, &calls)
	ctx, cancel := context.WithCancel(context.Background())
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	adapter, _ := testadapter.New(identity, func(request effectsv1.Request) effectexecution.Result {
		cancel()
		receipt, _ := effectsv1.NewSucceededReceipt(
			request, testReceiptID, "MAIL_SENT", nil,
			"result://mail/1", testDigest, "2026-07-15T10:00:01Z",
		)
		return definitiveResult(t, receipt)
	})
	registry, _ := effectexecution.NewRegistry([]effectexecution.Provider{adapter})
	finalizer := &externalFinalizerStub{calls: &calls}
	processor, _ := NewExternalProcessor(
		policy, &externalPlannerStub{spec: spec}, &preEffectStub{started: started},
		finalizer, registry,
	)
	if err := processor.process(ctx, asyncpostgres.JobClaim{}, 1); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(calls, "complete") {
		t.Fatalf("definitive response was not finalized: %v", calls)
	}
}

func TestStartedEffectReturnsDefensiveRequest(t *testing.T) {
	_, _, started := externalProcessorFixture(t, nil)
	first := started.Request()
	*first.TenantID = "99999999-9999-4999-8999-999999999999"
	second := started.Request()
	if second.TenantID == nil || *second.TenantID != testTenantID {
		t.Fatal("started effect exposed mutable request evidence")
	}
}
