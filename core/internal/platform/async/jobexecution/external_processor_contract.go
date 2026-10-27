package jobexecution

import (
	"context"
	"errors"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

const externalFinalizationTimeout = 5 * time.Second

const (
	codeEffectPlanInvalid        = "EFFECT_PLAN_INVALID"
	codeProviderNotRegistered    = "EFFECT_PROVIDER_NOT_REGISTERED"
	codeJobCancellationRequested = "JOB_CANCELLATION_REQUESTED"
)

var (
	ErrExternalProcessorDependencies = errors.New("external job processor dependencies are incomplete")
	ErrExternalProviderPanicked      = errors.New("external effect provider panicked")
	ErrExternalEffectStart           = errors.New("external effect start is invalid")
)

// ExternalPlanner creates immutable effect identity and request evidence only.
// It must not perform network I/O or execute the effect.
type ExternalPlanner interface {
	Plan(context.Context, asyncpostgres.JobClaim) (asyncpostgres.ExternalEffectRequestSpec, error)
}

// PreEffect opens a fresh TenantScope for every call. Start must perform the
// current authorization/cancellation check and bind MarkExternalEffectStarted
// in the same transaction. It returns only after an unambiguous commit.
// FailBeforeEffect performs the same fresh checks and binds FailJobBeforeEffect.
type PreEffect interface {
	Start(
		context.Context,
		asyncpostgres.JobClaim,
		asyncpostgres.ExternalEffectRequestSpec,
	) (StartedEffect, error)
	FailBeforeEffect(
		context.Context,
		asyncpostgres.JobClaim,
		asyncpostgres.JobFailure,
	) (asyncpostgres.JobFailureResult, error)
}

type ExternalFinalizer interface {
	CompleteExternalJob(
		context.Context, asyncpostgres.ExternalEffectPermit, effectsv1.Receipt,
	) (time.Time, error)
	FailExternalJob(
		context.Context, asyncpostgres.ExternalEffectPermit,
		effectsv1.Receipt, asyncpostgres.JobFailure,
	) (asyncpostgres.JobFailureResult, error)
	MarkExternalJobOutcomeUnknown(
		context.Context, asyncpostgres.ExternalEffectPermit,
	) (time.Time, error)
}

type effectInvocation interface {
	Invoke(context.Context, effectsv1.Request) (effectexecution.Result, error)
}

type effectResolver interface {
	Resolve(effectexecution.Identity) (effectInvocation, error)
}

var _ ExternalFinalizer = (asyncpostgres.JobLifecycleStore)(nil)
