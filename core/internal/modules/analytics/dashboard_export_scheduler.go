package analytics

import (
	"context"
	"crypto/sha256"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type dashboardExportScheduler struct{ effects *AsyncEffects }

func NewDashboardExportScheduler(
	catalogs platformasync.WriteCatalogs,
) (exportapp.Scheduler, error) {
	effects, err := NewAsyncEffects(catalogs)
	if err != nil {
		return nil, err
	}
	jobs, _, err := catalogs.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("resolve Analytics export job policies: %w", err)
	}
	for _, kind := range []string{DashboardExportRenderJobKind,
		DashboardExportDeleteJobKind, DashboardExportExpireJobKind} {
		policy, resolveErr := jobs.Resolve(kind, 1)
		if resolveErr != nil || !validDashboardExportJobPolicy(policy, kind) {
			return nil, fmt.Errorf("Analytics export job policy %s is invalid", kind)
		}
	}
	return &dashboardExportScheduler{effects: effects}, nil
}

func (scheduler *dashboardExportScheduler) Schedule(
	ctx context.Context, scope platformtx.TenantScope, job exportapp.Job,
) error {
	if scheduler == nil || scheduler.effects == nil || ctx == nil || ctx.Err() != nil ||
		job.JobID.Version() != 7 || job.ExportID.Version() != 7 || job.RunAt.IsZero() {
		return exportapp.ErrInvalid
	}
	kind := DashboardExportRenderJobKind
	if job.Action == exportapp.JobDelete {
		kind = DashboardExportDeleteJobKind
	} else if job.Action == exportapp.JobExpire {
		kind = DashboardExportExpireJobKind
	} else if job.Action != exportapp.JobRender {
		return exportapp.ErrInvalid
	}
	writer, authority, err := scheduler.effects.bind(scope)
	if err != nil {
		return fmt.Errorf("bind Analytics export scheduler: %w", err)
	}
	jobs, _, err := scheduler.effects.catalogs.Snapshot()
	if err != nil {
		return fmt.Errorf("snapshot Analytics export policy: %w", err)
	}
	policy, err := jobs.Resolve(kind, 1)
	if err != nil {
		return err
	}
	reference := job.ExportID.String()
	digest := sha256.Sum256([]byte(reference))
	tenantID := authority.ScopeID().String()
	effectKey := kind + ":" + reference
	enqueue, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: job.JobID.String(), Kind: kind, PayloadRevision: 1,
		TenantID: &tenantID, Authority: analyticsJobAuthority(authority),
		PayloadReference: reference, PayloadSHA256: fmt.Sprintf("%x", digest),
		RunAt: job.RunAt.UTC().Format(timeFormat), Priority: 100,
		DedupeKey: effectKey, EffectIdempotencyKey: &effectKey,
		MaximumAttempts: policy.MaximumAttempts,
	}, jobs)
	if err != nil {
		return err
	}
	result, err := writer.EnqueueJob(ctx, enqueue)
	if err != nil || !result.Created || result.ID != enqueue.JobID {
		if err != nil {
			return err
		}
		return exportapp.ErrUnavailable
	}
	return nil
}

const timeFormat = "2006-01-02T15:04:05.999999999Z07:00"

var _ exportapp.Scheduler = (*dashboardExportScheduler)(nil)
