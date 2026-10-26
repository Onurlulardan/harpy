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

func NewArchiveJobProcessor(release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore, effect ArchiveEffect,
	now func() time.Time,
) (jobexecution.Processor, error) {
	if store == nil || nilArchivePort(effect) || now == nil {
		return nil, ErrDocumentsRuntime
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read Documents archive catalog: %w", err)
	}
	policy, err := jobs.Resolve(asyncmanifestv1.DocumentsArchiveBuildJobKind, 1)
	if err != nil || policy.OwnerDomain != "documents" ||
		policy.EffectClass != jobsv1.EffectExternal ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, ErrDocumentsRuntime
	}
	provider, err := newArchiveProvider(effect, now)
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
	processor, err := jobexecution.NewExternalProcessor(policy, archivePlanner{},
		archivePreEffect{store: preEffect}, store, registry)
	if err != nil {
		return nil, fmt.Errorf("build Documents archive processor: %w", err)
	}
	return processor, nil
}
