package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const (
	iamSnapshotOwner                  = "iam"
	systemResourceLayoutSnapshotOwner = "system_resource_layout"
)

// IssueIAMSystemResources admits only the paired IAM system-resource snapshot.
// Both owners must be on the same cutover generation before one operation can
// bind their stores.
func (admission SnapshotOperationAdmission) IssueIAMSystemResources(
	verifiedGenerations map[string]int64,
) (transaction.VerifiedOperation, error) {
	iamGeneration := verifiedGenerations[iamSnapshotOwner]
	layoutGeneration := verifiedGenerations[systemResourceLayoutSnapshotOwner]
	if admission.Validate() != nil || len(verifiedGenerations) != 2 ||
		iamGeneration <= 0 || iamGeneration != layoutGeneration {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	operation, err := transaction.NewSnapshotAuthorityIssuer().Issue(
		snapshotReadOperationID,
		map[string]int64{
			iamSnapshotOwner:                  iamGeneration,
			systemResourceLayoutSnapshotOwner: layoutGeneration,
		},
	)
	if err != nil {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	return operation, nil
}
