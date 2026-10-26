package documents

import (
	"context"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *WorkerAsyncPort) EnqueueDeletion(ctx context.Context,
	scope platformtx.JobTenantScope, jobID, deletionID foundation.ID,
	payload []byte,
) error {
	authority, found := workerScopeAuthority(ctx, scope)
	if port == nil || !found || !jobID.UUIDv7() || !deletionID.UUIDv7() {
		return ErrDocumentsRuntime
	}
	reference, digest, err := jobcontract.Deletion(jobcontract.ID(scope.TenantID()),
		jobcontract.ID(deletionID))
	if err != nil || sha256Hex(payload) != digest {
		return ErrDocumentsRuntime
	}
	writer, err := asyncpostgres.BindJobWriter(ctx, scope, authority, port.catalogs)
	if err != nil {
		return ErrDocumentsRuntime
	}
	policy, err := port.jobs.Resolve(asyncmanifestv1.DocumentsBlobDeleteJobKind, 1)
	if err != nil || policy.OwnerDomain != "documents" {
		return ErrDocumentsRuntime
	}
	tenantID, key := scope.TenantID().String(),
		"documents:blob-delete:"+deletionID.String()
	job, err := jobsv1.NewEnqueue(jobsv1.Enqueue{JobID: jobID.String(),
		Kind: asyncmanifestv1.DocumentsBlobDeleteJobKind, PayloadRevision: 1,
		TenantID: &tenantID, Authority: authority, PayloadReference: reference,
		PayloadSHA256: digest, RunAt: port.clock.Now().UTC().Format(time.RFC3339Nano),
		Priority: 100, DedupeKey: key, EffectIdempotencyKey: &key,
		MaximumAttempts: policy.MaximumAttempts}, port.jobs)
	if err != nil {
		return ErrDocumentsRuntime
	}
	result, err := writer.EnqueueJob(ctx, job)
	if err != nil || result.ID != jobID.String() {
		return ErrDocumentsRuntime
	}
	return nil
}
