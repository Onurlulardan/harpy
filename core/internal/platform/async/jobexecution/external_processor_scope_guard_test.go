package jobexecution

import (
	"context"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type guardedResolver struct{ invocation effectInvocation }

func (resolver guardedResolver) Resolve(
	effectexecution.Identity,
) (effectInvocation, error) {
	return resolver.invocation, nil
}

type guardedInvocation struct{}

func (guardedInvocation) Invoke(
	context.Context,
	effectsv1.Request,
) (effectexecution.Result, error) {
	return effectexecution.Result{}, effectexecution.ErrNetworkInsideScope
}

func TestExternalProcessorTreatsTenantScopeNetworkGuardAsUnknown(t *testing.T) {
	calls := []string{}
	policy, spec, started := externalProcessorFixture(t, &calls)
	processor, err := newExternalProcessor(
		policy, &externalPlannerStub{spec: spec, calls: &calls},
		&preEffectStub{started: started, calls: &calls},
		&externalFinalizerStub{calls: &calls},
		guardedResolver{invocation: guardedInvocation{}}, &recordingJitter{},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = processor.process(context.Background(), asyncpostgres.JobClaim{}, 1)
	if err != nil {
		t.Fatalf("network guard error=%v", err)
	}
	assertOnlyUnknownFinalization(t, calls)
}
