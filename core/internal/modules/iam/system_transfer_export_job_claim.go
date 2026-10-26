package iam

import (
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

const systemExportEffectPrefix = "iam-system-export:"

type systemTransferExportClaimView interface {
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

func decodeSystemTransferExportClaim(
	claim systemTransferExportClaimView,
	policy jobsv1.KindPolicy,
) (systemtransfer.ExportJobClaim, error) {
	if claim == nil || ValidateSystemExportJobPolicy(policy) != nil {
		return systemtransfer.ExportJobClaim{}, ErrSystemTransferJobPolicy
	}
	tenantID, tenantScoped := claim.TenantID()
	payload, err := DecodeSystemTransferJobPayload(
		policy, claim.PayloadReference(), claim.PayloadSHA256(),
	)
	resource, resourceErr := systemTransferResource(payload.Resource)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	if err != nil || resourceErr != nil || !tenantScoped || tenantID == uuid.Nil ||
		!hasEffectKey || effectKey != systemExportEffectPrefix+payload.TransferID.String() ||
		claim.Kind() != policy.Kind || claim.PayloadRevision() != policy.PayloadRevision ||
		claim.EffectClass() != policy.EffectClass ||
		claim.MaximumAttempts() != policy.MaximumAttempts || claim.Attempt() < 1 ||
		claim.Attempt() > policy.MaximumAttempts ||
		claim.LeaseGeneration() != int64(claim.Attempt()) {
		return systemtransfer.ExportJobClaim{}, ErrSystemTransferJobPolicy
	}
	decoded := systemtransfer.ExportJobClaim{
		TenantID: tenantID, JobID: claim.JobID(), ExportID: payload.TransferID,
		AttemptID: claim.AttemptID(), LeaseOwnerID: claim.LeaseOwnerID(),
		Resource: resource, AttemptNumber: claim.Attempt(),
		MaximumAttempts: claim.MaximumAttempts(), LeaseGeneration: claim.LeaseGeneration(),
	}
	if decoded.Validate() != nil {
		return systemtransfer.ExportJobClaim{}, ErrSystemTransferJobPolicy
	}
	return decoded, nil
}

var _ systemTransferExportClaimView = asyncpostgres.JobClaim{}
