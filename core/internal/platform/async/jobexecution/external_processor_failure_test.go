package jobexecution

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution/testadapter"
)

func TestExternalProcessorFailsMissingProviderBeforeEffectStart(t *testing.T) {
	calls := []string{}
	policy, spec, started := externalProcessorFixture(t, &calls)
	registry, _ := effectexecution.NewRegistry(nil)
	processor, _ := NewExternalProcessor(
		policy, &externalPlannerStub{spec: spec, calls: &calls},
		&preEffectStub{started: started, calls: &calls},
		&externalFinalizerStub{calls: &calls}, registry,
	)
	if err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(calls, []string{"plan", "fail-before"}) {
		t.Fatalf("missing provider crossed effect boundary: %v", calls)
	}
}

func TestExternalProcessorPersistsPlannerErrorBeforeEffectStart(t *testing.T) {
	calls := []string{}
	policy, _, started := externalProcessorFixture(t, &calls)
	registry, _ := effectexecution.NewRegistry(nil)
	processor, _ := NewExternalProcessor(
		policy, &externalPlannerStub{err: errors.New("invalid pure plan"), calls: &calls},
		&preEffectStub{started: started, calls: &calls},
		&externalFinalizerStub{calls: &calls}, registry,
	)
	if err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(calls, []string{"plan", "fail-before"}) {
		t.Fatalf("planner error crossed effect boundary: %v", calls)
	}
}

func TestExternalProcessorPersistsDefinitiveFailureWithFullJitter(t *testing.T) {
	calls := []string{}
	policy, spec, started := externalProcessorFixture(t, &calls)
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	adapter, _ := testadapter.New(identity, func(request effectsv1.Request) effectexecution.Result {
		receipt, err := effectsv1.NewFailedReceipt(
			request, testReceiptID, "MAIL_RATE_LIMITED", effectsv1.ErrorTransient,
			nil, "2026-07-15T10:00:01Z",
		)
		if err != nil {
			t.Fatal(err)
		}
		return definitiveResult(t, receipt)
	})
	registry, _ := effectexecution.NewRegistry([]effectexecution.Provider{adapter})
	jitter := &recordingJitter{value: 7}
	finalizer := &externalFinalizerStub{calls: &calls}
	processor, _ := newExternalProcessor(
		policy, &externalPlannerStub{spec: spec}, &preEffectStub{started: started},
		finalizer, registryResolver{registry: registry}, jitter,
	)
	if err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1); err != nil {
		t.Fatal(err)
	}
	wantBound := uint64(5*time.Second) + 1
	if jitter.bound != wantBound || !slices.Contains(calls, "fail") ||
		finalizer.receipt.ErrorClass == nil ||
		*finalizer.receipt.ErrorClass != effectsv1.ErrorTransient {
		t.Fatalf("bound=%d calls=%v receipt=%+v", jitter.bound, calls, finalizer.receipt)
	}
}

func TestExternalProcessorDurablyCancelsBeforeEffect(t *testing.T) {
	calls := []string{}
	policy, spec, started := externalProcessorFixture(t, &calls)
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	adapter, _ := testadapter.New(identity, func(effectsv1.Request) effectexecution.Result {
		t.Fatal("provider was invoked after cancellation")
		return effectexecution.Unknown()
	})
	registry, _ := effectexecution.NewRegistry([]effectexecution.Provider{adapter})
	processor, _ := NewExternalProcessor(
		policy, &externalPlannerStub{spec: spec, calls: &calls},
		&preEffectStub{
			started: started, startErr: asyncpostgres.ErrJobNotControllable, calls: &calls,
		}, &externalFinalizerStub{calls: &calls}, registry,
	)
	if err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(calls, []string{"plan", "start", "fail-before"}) {
		t.Fatalf("cancellation flow=%v", calls)
	}
}

func TestExternalProcessorRequiresExternalPolicy(t *testing.T) {
	policy, _ := jobsv1.NewKindPolicy(
		"notification.email", 1, "notifications", jobsv1.EffectDatabaseTransactional, 3,
	)
	registry, _ := effectexecution.NewRegistry(nil)
	if _, err := NewExternalProcessor(
		policy, &externalPlannerStub{}, &preEffectStub{},
		&externalFinalizerStub{}, registry,
	); err != ErrExternalProcessorDependencies {
		t.Fatalf("database policy error=%v", err)
	}
}

func TestExternalProcessorRejectsTypedNilDependencies(t *testing.T) {
	policy, _, _ := externalProcessorFixture(t, nil)
	registry, _ := effectexecution.NewRegistry(nil)
	var planner *externalPlannerStub
	if _, err := NewExternalProcessor(
		policy, planner, &preEffectStub{},
		&externalFinalizerStub{}, registry,
	); err != ErrExternalProcessorDependencies {
		t.Fatalf("typed nil planner error=%v", err)
	}
}
