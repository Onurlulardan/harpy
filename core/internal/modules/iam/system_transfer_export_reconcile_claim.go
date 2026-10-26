package iam

import (
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

const systemExportReconcileEffectPrefix = "iam-system-export-reconcile:"

type systemTransferExportReconcileClaimView interface {
	AttemptID() uuid.UUID
	JobID() uuid.UUID
	TenantID() (uuid.UUID, bool)
	Kind() string
	PayloadRevision() int
	PayloadReference() string
	PayloadSHA256() string
	EffectClass() jobsv1.EffectClass
	EffectIdempotencyKey() (string, bool)
	Attempt() int
	MaximumAttempts() int
	LeaseOwnerID() uuid.UUID
	LeaseGeneration() int64
}

func decodeSystemTransferExportReconcileClaim(
	claim systemTransferExportReconcileClaimView,
	policy jobsv1.KindPolicy,
) (systemtransfer.ExportReconciliationClaim, error) {
	if claim == nil || ValidateSystemExportReconcileJobPolicy(policy) != nil {
		return systemtransfer.ExportReconciliationClaim{}, ErrSystemTransferJobPolicy
	}
	tenantID, tenantScoped := claim.TenantID()
	payload, err := DecodeSystemTransferJobPayload(
		policy, claim.PayloadReference(), claim.PayloadSHA256())
	resource, resourceErr := systemTransferResource(payload.Resource)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	if err != nil || resourceErr != nil || !tenantScoped || tenantID == uuid.Nil || !hasEffectKey ||
		effectKey != systemExportReconcileEffectPrefix+payload.TransferID.String() ||
		claim.Kind() != policy.Kind || claim.PayloadRevision() != policy.PayloadRevision ||
		claim.EffectClass() != policy.EffectClass ||
		claim.MaximumAttempts() != policy.MaximumAttempts || claim.Attempt() < 1 ||
		claim.Attempt() > policy.MaximumAttempts ||
		claim.LeaseGeneration() != int64(claim.Attempt()) {
		return systemtransfer.ExportReconciliationClaim{}, ErrSystemTransferJobPolicy
	}
	decoded := systemtransfer.ExportReconciliationClaim{
		TenantID: tenantID, ExportID: payload.TransferID, JobID: claim.JobID(),
		AttemptID: claim.AttemptID(), LeaseOwnerID: claim.LeaseOwnerID(),
		Resource: resource, Attempt: claim.Attempt(),
		MaximumAttempts: claim.MaximumAttempts(),
		LeaseGeneration: claim.LeaseGeneration(),
	}
	if decoded.Validate() != nil {
		return systemtransfer.ExportReconciliationClaim{}, ErrSystemTransferJobPolicy
	}
	return decoded, nil
}

var _ systemTransferExportReconcileClaimView = asyncpostgres.JobClaim{}
