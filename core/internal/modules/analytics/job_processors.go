package analytics

import (
	"context"
	"fmt"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	jobspostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/jobs"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func NewJobProcessors(
	release platformasync.ReleaseCatalogs,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read Analytics job catalog: %w", err)
	}
	kinds := []string{
		asyncmanifestv1.AnalyticsQueryRevalidateJobKind,
		asyncmanifestv1.AnalyticsCacheWarmJobKind,
		asyncmanifestv1.AnalyticsPublicationRepairJobKind,
		asyncmanifestv1.AnalyticsRetentionJobKind,
	}
	processors := make([]jobexecution.Processor, 0, len(kinds))
	for _, kind := range kinds {
		policy, err := analyticsJobPolicy(jobs, kind)
		if err != nil {
			return nil, err
		}
		binding, err := analyticsJobBinding(kind)
		if err != nil {
			return nil, err
		}
		processor, err := jobexecution.NewDatabaseProcessor(policy, store, binding)
		if err != nil {
			return nil, fmt.Errorf("build %s processor: %w", kind, err)
		}
		processors = append(processors, processor)
	}
	return processors, nil
}

func analyticsJobPolicy(jobs jobsv1.Catalog, kind string) (jobsv1.KindPolicy, error) {
	policy, err := jobs.Resolve(kind, 1)
	if err != nil || policy.OwnerDomain != "analytics" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return jobsv1.KindPolicy{}, fmt.Errorf("Analytics job policy %s is invalid", kind)
	}
	return policy, nil
}

func analyticsJobBinding(kind string) (platformasync.DatabaseJobBinding, error) {
	if kind == asyncmanifestv1.AnalyticsRetentionJobKind {
		return newAnalyticsRetentionBinding()
	}
	switch kind {
	case asyncmanifestv1.AnalyticsQueryRevalidateJobKind,
		asyncmanifestv1.AnalyticsCacheWarmJobKind,
		asyncmanifestv1.AnalyticsPublicationRepairJobKind:
	default:
		return platformasync.DatabaseJobBinding{}, fmt.Errorf(
			"Analytics job kind %s is unsupported", kind)
	}
	return platformasync.NewDatabaseJobBinding("analytics",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			store := jobspostgres.NewStore(func(
				ctx context.Context, query string, arguments ...any,
			) jobspostgres.Row {
				return database.QueryRow(ctx, query, arguments...)
			})
			switch kind {
			case asyncmanifestv1.AnalyticsQueryRevalidateJobKind:
				return &analyticsQueryRevalidateCommand{store: store}
			case asyncmanifestv1.AnalyticsCacheWarmJobKind:
				return &analyticsCacheWarmCommand{store: store}
			case asyncmanifestv1.AnalyticsPublicationRepairJobKind:
				return &analyticsPublicationRepairCommand{store: store, clock: time.Now}
			default:
				return nil
			}
		})
}
