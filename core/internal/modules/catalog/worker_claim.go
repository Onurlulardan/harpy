package catalog

import (
	"context"
	"encoding/hex"
	"strings"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (processor *workerProcessor) claimSnapshot(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) (lifecyclepostgres.JobSnapshot, error) {
	tenantID, tenantScoped := claim.TenantID()
	prefix := catalogOperationReference(tenantID, "")
	operationID, parseErr := uuid.Parse(strings.TrimPrefix(claim.PayloadReference(), prefix))
	digest, digestErr := hex.DecodeString(claim.PayloadSHA256())
	if processor == nil || ctx == nil || ctx.Err() != nil || !tenantScoped ||
		tenantID == uuid.Nil || claim.JobID().Version() != uuid.Version(7) ||
		claim.AttemptID().Version() != uuid.Version(7) ||
		claim.Kind() != asyncmanifestv1.CatalogOperationResumeJobKind ||
		claim.PayloadRevision() != processor.policy.PayloadRevision ||
		claim.EffectClass() != processor.policy.EffectClass ||
		claim.MaximumAttempts() != processor.policy.MaximumAttempts ||
		parseErr != nil || operationID.Version() != uuid.Version(7) ||
		claim.PayloadReference() != catalogOperationReference(tenantID, operationID.String()) ||
		digestErr != nil || len(digest) != 32 {
		return lifecyclepostgres.JobSnapshot{}, asyncUnavailable()
	}
	var snapshot lifecyclepostgres.JobSnapshot
	err := processor.scopes.InJobTenant(ctx, tenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			value, readErr := processor.claims.Snapshot(scoped, scope, operationID)
			if readErr != nil {
				return readErr
			}
			snapshot = value
			return nil
		})
	if err != nil || !snapshot.Operation.Valid() || snapshot.ActorID == uuid.Nil ||
		snapshot.Operation.ID.String() != operationID.String() ||
		snapshot.Operation.Plan.SHA256 != claim.PayloadSHA256() {
		return lifecyclepostgres.JobSnapshot{}, asyncUnavailable()
	}
	return snapshot, nil
}
