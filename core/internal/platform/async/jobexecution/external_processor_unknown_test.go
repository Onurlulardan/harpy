package jobexecution

import (
	"context"
	"errors"
	"slices"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution/testadapter"
)

func TestExternalProcessorMarksExplicitUnknownWithoutRetry(t *testing.T) {
	calls := []string{}
	processor := externalProcessorWithHandler(t, &calls, func(effectsv1.Request) effectexecution.Result {
		return effectexecution.Unknown()
	})
	if err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1); err != nil {
		t.Fatal(err)
	}
	assertOnlyUnknownFinalization(t, calls)
}

func TestExternalProcessorMarksProviderPanicUnknown(t *testing.T) {
	calls := []string{}
	processor := externalProcessorWithHandler(t, &calls, func(effectsv1.Request) effectexecution.Result {
		panic("provider secret must not escape")
	})
	err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1)
	if err != nil {
		t.Fatalf("panic error=%v", err)
	}
	assertOnlyUnknownFinalization(t, calls)
}

func TestExternalProcessorMarksMalformedProviderResultUnknown(t *testing.T) {
	calls := []string{}
	processor := externalProcessorWithHandler(t, &calls, func(effectsv1.Request) effectexecution.Result {
		return effectexecution.Result{}
	})
	err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1)
	if err != nil {
		t.Fatalf("malformed result error=%v", err)
	}
	assertOnlyUnknownFinalization(t, calls)
}

func TestExternalProcessorSurfacesUnknownPersistenceFailure(t *testing.T) {
	calls := []string{}
	processor := externalProcessorWithHandler(t, &calls, func(effectsv1.Request) effectexecution.Result {
		return effectexecution.Unknown()
	})
	processor.finalizer = &externalFinalizerStub{
		calls: &calls, err: errors.New("database unavailable"),
	}
	if err := processor.process(context.Background(), asyncpostgres.JobClaim{}, 1); err == nil {
		t.Fatal("unknown persistence failure was hidden")
	}
	assertOnlyUnknownFinalization(t, calls)
}

func externalProcessorWithHandler(
	t *testing.T,
	calls *[]string,
	handler testadapter.Handler,
) *ExternalProcessor {
	t.Helper()
	policy, spec, started := externalProcessorFixture(t, calls)
	identity, _ := effectexecution.NewIdentity("test.mail", "send")
	adapter, _ := testadapter.New(identity, handler)
	registry, _ := effectexecution.NewRegistry([]effectexecution.Provider{adapter})
	processor, err := NewExternalProcessor(
		policy, &externalPlannerStub{spec: spec, calls: calls},
		&preEffectStub{started: started, calls: calls},
		&externalFinalizerStub{calls: calls}, registry,
	)
	if err != nil {
		t.Fatal(err)
	}
	return processor
}

func assertOnlyUnknownFinalization(t *testing.T, calls []string) {
	t.Helper()
	if !slices.Contains(calls, "unknown") || slices.Contains(calls, "fail") ||
		slices.Contains(calls, "complete") || slices.Contains(calls, "fail-before") {
		t.Fatalf("unknown outcome was retried or mis-finalized: %v", calls)
	}
}
