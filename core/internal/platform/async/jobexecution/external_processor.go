package jobexecution

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (processor *ExternalProcessor) Process(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	return processor.process(ctx, claim, claim.Attempt())
}

func (processor *ExternalProcessor) process(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	attempt int,
) error {
	if processor == nil || ctx == nil {
		return ErrExternalProcessorDependencies
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	spec, err := processor.planner.Plan(ctx, claim)
	if err != nil {
		return processor.failBeforeEffect(
			ctx, claim, codeEffectPlanInvalid,
			stableFailureDigest(
				codeEffectPlanInvalid, processor.policy.Kind,
				strconv.Itoa(processor.policy.PayloadRevision),
			),
		)
	}
	identity, err := effectexecution.NewIdentity(spec.Provider(), spec.Operation())
	if err != nil {
		return processor.failBeforeEffect(
			ctx, claim, codeEffectPlanInvalid,
			stableFailureDigest(codeEffectPlanInvalid, spec.Provider(), spec.Operation()),
		)
	}
	invocation, err := processor.effects.Resolve(identity)
	if err != nil {
		if errors.Is(err, effectexecution.ErrProviderNotRegistered) {
			return processor.failBeforeEffect(
				ctx, claim, codeProviderNotRegistered,
				stableFailureDigest(codeProviderNotRegistered, identity.Provider(), identity.Operation()),
			)
		}
		return fmt.Errorf("resolve external effect provider: %w", err)
	}
	started, err := processor.preEffect.Start(ctx, claim, spec)
	if err != nil {
		if errors.Is(err, asyncpostgres.ErrJobNotControllable) {
			return processor.cancelBeforeEffect(ctx, claim)
		}
		return fmt.Errorf("start external job effect: %w", err)
	}
	if !started.validFor(spec) {
		return errors.Join(ErrExternalEffectStart, processor.markUnknown(ctx, started))
	}
	result, invokeErr := invokeExternal(ctx, invocation, started.request)
	if invokeErr != nil {
		return processor.markUnknown(ctx, started)
	}
	return processor.finalizeResult(ctx, claim, attempt, started, result)
}

func invokeExternal(
	ctx context.Context,
	invocation effectInvocation,
	request effectsv1.Request,
) (result effectexecution.Result, err error) {
	defer func() {
		if recover() != nil {
			result = effectexecution.Result{}
			err = ErrExternalProviderPanicked
		}
	}()
	return invocation.Invoke(ctx, request)
}
