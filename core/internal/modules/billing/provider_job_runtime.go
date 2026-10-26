package billing

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type ProviderJobRuntime struct {
	registry providerapp.Registry
}

func NewProviderJobRuntime(
	registry providerapp.Registry,
) (*ProviderJobRuntime, error) {
	if len(registry.CommandProviders()) == 0 ||
		len(registry.ReconciliationProviders()) == 0 {
		return nil, providerapp.ErrConfiguration
	}
	return &ProviderJobRuntime{registry: registry}, nil
}

func (runtime *ProviderJobRuntime) Processors(
	catalog jobsv1.Catalog,
	store asyncpostgres.WorkerStore,
) ([]jobexecution.Processor, error) {
	if runtime == nil || store == nil {
		return nil, providerapp.ErrConfiguration
	}
	commandPolicy, err := providerJobPolicy(
		catalog, providerapp.ProviderCommandJobKind,
	)
	if err != nil {
		return nil, err
	}
	reconcilePolicy, err := providerJobPolicy(
		catalog, providerapp.ProviderReconciliationJobKind,
	)
	if err != nil {
		return nil, err
	}
	preEffect, err := asyncpostgres.BindExternalPreEffect(store)
	if err != nil {
		return nil, fmt.Errorf("bind billing provider pre-effect: %w", err)
	}
	effects, err := runtime.effects()
	if err != nil {
		return nil, err
	}
	command, err := jobexecution.NewExternalProcessor(
		commandPolicy, providerCommandPlanner(),
		providerJobPreEffect{store: preEffect}, store, effects,
	)
	if err != nil {
		return nil, fmt.Errorf("build billing provider command processor: %w", err)
	}
	reconcile, err := jobexecution.NewExternalProcessor(
		reconcilePolicy, providerReconciliationPlanner(),
		providerJobPreEffect{store: preEffect}, store, effects,
	)
	if err != nil {
		return nil, fmt.Errorf("build billing provider reconcile processor: %w", err)
	}
	finalizePolicy, err := providerFinalizePolicy(catalog)
	if err != nil {
		return nil, err
	}
	finalizeBinding, err := newProviderFinalizeJobBinding()
	if err != nil {
		return nil, err
	}
	finalize, err := jobexecution.NewDatabaseProcessor(
		finalizePolicy, store, finalizeBinding,
	)
	if err != nil {
		return nil, fmt.Errorf("build billing provider finalizer processor: %w", err)
	}
	return []jobexecution.Processor{command, reconcile, finalize}, nil
}

func providerJobPolicy(catalog jobsv1.Catalog, kind string) (jobsv1.KindPolicy, error) {
	policy, err := catalog.Resolve(kind, 1)
	if err != nil || policy.OwnerDomain != "billing" ||
		policy.EffectClass != jobsv1.EffectExternal ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return jobsv1.KindPolicy{}, fmt.Errorf("billing provider job policy is invalid")
	}
	return policy, nil
}

func providerFinalizePolicy(catalog jobsv1.Catalog) (jobsv1.KindPolicy, error) {
	policy, err := catalog.Resolve(providerapp.ProviderFinalizeJobKind, 1)
	if err != nil || policy.OwnerDomain != "billing" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return jobsv1.KindPolicy{},
			fmt.Errorf("billing provider finalizer job policy is invalid")
	}
	return policy, nil
}
