package iam

import (
	"crypto/subtle"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type SystemTransferJobPayload struct {
	Resource   SystemTransferResource
	TransferID uuid.UUID
}

func DecodeSystemTransferJobPayload(
	policy jobsv1.KindPolicy,
	reference string,
	digest string,
) (SystemTransferJobPayload, error) {
	var policyErr error
	switch policy.Kind {
	case SystemExportJobKind:
		policyErr = ValidateSystemExportJobPolicy(policy)
	case SystemImportJobKind:
		policyErr = ValidateSystemImportJobPolicy(policy)
	case SystemExportReconcileJobKind:
		policyErr = ValidateSystemExportReconcileJobPolicy(policy)
	default:
		policyErr = ErrSystemTransferJobPolicy
	}
	if policyErr != nil {
		return SystemTransferJobPayload{}, policyErr
	}
	resource, transferID, err := systemtransfer.ParseJobReference(reference)
	if err != nil {
		return SystemTransferJobPayload{}, ErrSystemTransferJobPolicy
	}
	wantedDigest, err := systemtransfer.JobReferenceSHA256(reference)
	if err != nil || len(digest) != len(wantedDigest) ||
		subtle.ConstantTimeCompare([]byte(digest), []byte(wantedDigest)) != 1 {
		return SystemTransferJobPayload{}, ErrSystemTransferJobPolicy
	}
	rootResource, err := ParseSystemTransferResource(string(resource))
	if err != nil {
		return SystemTransferJobPayload{}, ErrSystemTransferJobPolicy
	}
	return SystemTransferJobPayload{Resource: rootResource, TransferID: transferID}, nil
}
