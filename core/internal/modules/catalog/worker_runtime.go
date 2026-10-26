package catalog

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	registrypostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/registry"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type WorkerScopeCoordinator interface {
	InJobTenant(context.Context, uuid.UUID,
		func(context.Context, platformtx.JobTenantScope) error) error
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type PlatformWorkerDependencies struct {
	Release     platformasync.ReleaseCatalogs
	Catalogs    platformasync.WriteCatalogs
	Generations HTTPGenerationResolver
	Scopes      WorkerScopeCoordinator
	Keys        lifecycleapp.KeyResolver
	Owners      ComponentOwnerDirectory
	Demo        lifecycleapp.DemoOwner
}

type workerProcessor struct {
	policy      jobsv1.KindPolicy
	generations HTTPGenerationResolver
	scopes      WorkerScopeCoordinator
	claims      lifecyclepostgres.JobBinding
	lifecycle   *lifecycleapp.Application
	async       *AsyncPort
	values      runtimeValues
}

func NewPlatformJobProcessor(
	dependencies PlatformWorkerDependencies,
) (jobexecution.Processor, error) {
	if nilHTTPPort(dependencies.Generations) || nilHTTPPort(dependencies.Scopes) ||
		nilHTTPPort(dependencies.Keys) || nilHTTPPort(dependencies.Owners) ||
		nilHTTPPort(dependencies.Demo) {
		return nil, fmt.Errorf("Catalog worker dependencies are required")
	}
	jobs, _, err := dependencies.Release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("load Catalog worker policy: %w", err)
	}
	policy, err := jobs.Resolve(asyncmanifestv1.CatalogOperationResumeJobKind, 1)
	if err != nil || policy.OwnerDomain != "catalog" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, fmt.Errorf("Catalog worker policy is unavailable")
	}
	values := runtimeValues{}
	async, err := NewAsyncPort(dependencies.Catalogs, values)
	if err != nil {
		return nil, err
	}
	registry := registrypostgres.NewBinder()
	lifecycle, err := lifecycleapp.NewApplication(lifecycleapp.Dependencies{
		Binder: lifecyclepostgres.NewBinder(), Registry: registry, Keys: dependencies.Keys,
		Owners: dependencies.Owners, Authorizer: lifecycleAuthorizer{}, Jobs: async, Events: async,
		Backup: NewBackupVerifier(), Demo: dependencies.Demo,
		IDs: values, Clock: values,
	})
	if err != nil {
		return nil, fmt.Errorf("build Catalog worker lifecycle: %w", err)
	}
	claims, err := lifecyclepostgres.NewJobBinding()
	if err != nil {
		return nil, fmt.Errorf("build Catalog worker claim binding: %w", err)
	}
	return &workerProcessor{policy: policy, generations: dependencies.Generations,
		scopes: dependencies.Scopes, claims: claims, lifecycle: lifecycle,
		async: async, values: values}, nil
}

func (processor *workerProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

var _ jobexecution.Processor = (*workerProcessor)(nil)
