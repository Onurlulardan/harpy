package iam

import (
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

type systemTransferImportClaimView interface {
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

func decodeSystemTransferImportClaim(
	claim systemTransferImportClaimView,
	policy jobsv1.KindPolicy,
) (systemtransfer.AttemptClaim, error) {
	if claim == nil || ValidateSystemImportJobPolicy(policy) != nil {
		return systemtransfer.AttemptClaim{}, ErrSystemTransferJobPolicy
	}
	tenantID, tenantScoped := claim.TenantID()
	payload, err := DecodeSystemTransferJobPayload(
		policy, claim.PayloadReference(), claim.PayloadSHA256(),
	)
	resource, resourceErr := publicresource.ParseResource(string(payload.Resource))
	_, hasEffectKey := claim.EffectIdempotencyKey()
	if err != nil || resourceErr == false || !tenantScoped || tenantID == uuid.Nil ||
		claim.Kind() != policy.Kind || claim.PayloadRevision() != policy.PayloadRevision ||
		claim.EffectClass() != policy.EffectClass || hasEffectKey ||
		claim.MaximumAttempts() != policy.MaximumAttempts || claim.Attempt() < 1 ||
		claim.Attempt() > policy.MaximumAttempts ||
		claim.LeaseGeneration() != int64(claim.Attempt()) {
		return systemtransfer.AttemptClaim{}, ErrSystemTransferJobPolicy
	}
	decoded := systemtransfer.AttemptClaim{
		TenantID: tenantID, ImportID: payload.TransferID, JobID: claim.JobID(),
		AttemptID: claim.AttemptID(), LeaseOwnerID: claim.LeaseOwnerID(),
		LeaseGeneration: claim.LeaseGeneration(), Resource: resource,
	}
	if decoded.Validate() != nil {
		return systemtransfer.AttemptClaim{}, ErrSystemTransferJobPolicy
	}
	return decoded, nil
}

var _ systemTransferImportClaimView = asyncpostgres.JobClaim{}
