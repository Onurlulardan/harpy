package catalog

import (
	"context"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *AsyncPort) Dispatch(
	ctx context.Context, scope platformtx.TenantScope, operation catalogapp.Operation,
) error {
	if !operation.Valid() {
		return catalogapp.ErrInvalid
	}
	writer, authority, err := port.bindAsync(ctx, scope)
	if err != nil {
		return err
	}
	policy, err := port.jobs.Resolve(asyncmanifestv1.CatalogOperationResumeJobKind, 1)
	if err != nil {
		return asyncUnavailable()
	}
	tenantID := scope.TenantID().String()
	reference := catalogOperationReference(scope.TenantID(), operation.ID.String())
	job, err := jobsv1.NewEnqueue(jobsv1.Enqueue{JobID: operation.ID.String(),
		Kind: asyncmanifestv1.CatalogOperationResumeJobKind, PayloadRevision: 1,
		TenantID: &tenantID, Authority: catalogJobAuthority(authority), PayloadReference: reference,
		PayloadSHA256: operation.Plan.SHA256, RunAt: operation.StartedAt.Format(time.RFC3339Nano),
		Priority: 100, DedupeKey: operation.ID.String(), MaximumAttempts: policy.MaximumAttempts}, port.jobs)
	if err != nil {
		return asyncUnavailable()
	}
	result, err := writer.EnqueueJob(ctx, job)
	if err != nil || result.ID != operation.ID.String() {
		return asyncUnavailable()
	}
	return nil
}
