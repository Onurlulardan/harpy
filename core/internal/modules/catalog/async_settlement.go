package catalog

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (port *AsyncPort) Settle(
	ctx context.Context,
	scope platformtx.TenantScope,
	claim asyncpostgres.JobClaim,
	operation catalogapp.Operation,
	nextID uuid.UUID,
	runAt time.Time,
) error {
	tenantID, tenantScoped := claim.TenantID()
	if port == nil || !tenantScoped || tenantID != scope.TenantID() ||
		claim.Kind() != asyncmanifestv1.CatalogOperationResumeJobKind ||
		claim.PayloadRevision() != 1 || !operation.Valid() ||
		claim.PayloadSHA256() != operation.Plan.SHA256 {
		return asyncUnavailable()
	}
	writer, authority, err := port.bindAsync(ctx, scope)
	if err != nil {
		return err
	}
	if !catalogapp.IsTerminalResume(operation) {
		if nextID == uuid.Nil || nextID.Version() != uuid.Version(7) ||
			nextID == claim.JobID() || runAt.IsZero() || runAt.Location() != time.UTC ||
			runAt.Nanosecond()%int(time.Microsecond) != 0 {
			return asyncUnavailable()
		}
		policy, policyErr := port.jobs.Resolve(
			asyncmanifestv1.CatalogOperationResumeJobKind, 1,
		)
		if policyErr != nil {
			return asyncUnavailable()
		}
		tenant := tenantID.String()
		reference := catalogOperationReference(tenantID, operation.ID.String())
		job, jobErr := jobsv1.NewEnqueue(jobsv1.Enqueue{
			JobID: nextID.String(), Kind: policy.Kind, PayloadRevision: policy.PayloadRevision,
			TenantID: &tenant, Authority: catalogJobAuthority(authority),
			PayloadReference: reference, PayloadSHA256: operation.Plan.SHA256,
			RunAt: runAt.Format(time.RFC3339Nano), Priority: int32(claim.Priority()),
			DedupeKey: operation.ID.String() + ":resume:" +
				strconv.FormatUint(operation.Revision, 10),
			MaximumAttempts: policy.MaximumAttempts,
		}, port.jobs)
		if jobErr != nil {
			return asyncUnavailable()
		}
		result, enqueueErr := writer.EnqueueJob(ctx, job)
		if enqueueErr != nil || result.ID != nextID.String() {
			return asyncUnavailable()
		}
	}
	completedAt, err := writer.CompleteDatabaseJob(ctx, claim)
	if errors.Is(err, asyncpostgres.ErrJobLeaseLost) {
		return err
	}
	if err != nil || completedAt.IsZero() {
		return asyncUnavailable()
	}
	return nil
}

func catalogOperationReference(tenantID uuid.UUID, operationID string) string {
	return fmt.Sprintf("catalog-operation://tenant/%s/operation/%s", tenantID, operationID)
}
