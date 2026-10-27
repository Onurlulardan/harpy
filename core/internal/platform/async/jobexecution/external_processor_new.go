package jobexecution

import (
	"reflect"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type ExternalProcessor struct {
	policy    jobsv1.KindPolicy
	planner   ExternalPlanner
	preEffect PreEffect
	finalizer ExternalFinalizer
	effects   effectResolver
	jitter    platformasync.JitterSource
}

var _ Processor = (*ExternalProcessor)(nil)

func NewExternalProcessor(
	policy jobsv1.KindPolicy,
	planner ExternalPlanner,
	preEffect PreEffect,
	finalizer ExternalFinalizer,
	effects effectexecution.Registry,
) (*ExternalProcessor, error) {
	return newExternalProcessor(
		policy, planner, preEffect, finalizer,
		registryResolver{registry: effects}, cryptographicJitter{},
	)
}

func newExternalProcessor(
	policy jobsv1.KindPolicy,
	planner ExternalPlanner,
	preEffect PreEffect,
	finalizer ExternalFinalizer,
	effects effectResolver,
	jitter platformasync.JitterSource,
) (*ExternalProcessor, error) {
	if policy.Validate() != nil || policy.EffectClass != jobsv1.EffectExternal ||
		missingExternalDependency(planner) || missingExternalDependency(preEffect) ||
		missingExternalDependency(finalizer) || missingExternalDependency(effects) ||
		missingExternalDependency(jitter) {
		return nil, ErrExternalProcessorDependencies
	}
	return &ExternalProcessor{
		policy: policy, planner: planner, preEffect: preEffect,
		finalizer: finalizer, effects: effects, jitter: jitter,
	}, nil
}

func missingExternalDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func (processor *ExternalProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

type registryResolver struct{ registry effectexecution.Registry }

func (resolver registryResolver) Resolve(
	identity effectexecution.Identity,
) (effectInvocation, error) {
	return resolver.registry.Resolve(identity)
}
