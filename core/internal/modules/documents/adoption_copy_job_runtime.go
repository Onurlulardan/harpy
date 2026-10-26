package documents

import (
	"fmt"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func NewAdoptionCopyJobProcessor(release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore, effect AdoptionCopyEffect,
	now func() time.Time,
) (jobexecution.Processor, error) {
	if store == nil || nilArchivePort(effect) || now == nil {
		return nil, ErrDocumentsRuntime
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read Documents adoption copy catalog: %w", err)
	}
	policy, err := jobs.Resolve(asyncmanifestv1.DocumentsAdoptionCopyJobKind, 1)
	if err != nil || policy.OwnerDomain != "documents" ||
		policy.EffectClass != jobsv1.EffectExternal ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, ErrDocumentsRuntime
	}
	provider, err := newAdoptionCopyProvider(effect, now)
	if err != nil {
		return nil, err
	}
	registry, err := effectexecution.NewRegistry([]effectexecution.Provider{provider})
	if err != nil {
		return nil, err
	}
	preEffect, err := asyncpostgres.BindExternalPreEffect(store)
	if err != nil {
		return nil, err
	}
	processor, err := jobexecution.NewExternalProcessor(policy, adoptionCopyPlanner{},
		adoptionCopyPreEffect{store: preEffect}, store, registry)
	if err != nil {
		return nil, fmt.Errorf("build Documents adoption copy processor: %w", err)
	}
	return processor, nil
}
