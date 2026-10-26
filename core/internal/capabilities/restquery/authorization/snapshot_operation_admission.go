package authorization

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// Record and aggregate QuerySpecs share this read-only snapshot operation ID;
// their different query semantics are already sealed inside resolved.QuerySpec.
const snapshotReadOperationID = "restquery.execute"

// SnapshotOperationAdmission is the sole REST Query admission boundary allowed
// to mint a multi-owner snapshot operation capability.
type SnapshotOperationAdmission struct {
	validated bool
}

func NewSnapshotOperationAdmission() SnapshotOperationAdmission {
	return SnapshotOperationAdmission{validated: true}
}

func (admission SnapshotOperationAdmission) Validate() error {
	if !admission.validated {
		return errSnapshotOperationAdmission
	}
	return nil
}

func (admission SnapshotOperationAdmission) Issue(
	spec resolved.QuerySpec,
	verifiedGenerations map[string]int64,
) (transaction.VerifiedOperation, error) {
	if admission.Validate() != nil {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	generations, err := exactSnapshotGenerations(spec, verifiedGenerations)
	if err != nil {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	operation, err := transaction.NewSnapshotAuthorityIssuer().Issue(
		snapshotReadOperationID, generations,
	)
	if err != nil {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	return operation, nil
}

func (admission SnapshotOperationAdmission) IssueIAM(
	verifiedGenerations map[string]int64,
) (transaction.VerifiedOperation, error) {
	if admission.Validate() != nil || len(verifiedGenerations) != 1 ||
		verifiedGenerations["iam"] <= 0 {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	operation, err := transaction.NewSnapshotAuthorityIssuer().Issue(
		snapshotReadOperationID, map[string]int64{"iam": verifiedGenerations["iam"]},
	)
	if err != nil {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	return operation, nil
}

func (admission SnapshotOperationAdmission) IssueForms(
	verifiedGenerations map[string]int64,
) (transaction.VerifiedOperation, error) {
	if admission.Validate() != nil || len(verifiedGenerations) != 1 ||
		verifiedGenerations["forms"] <= 0 {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	operation, err := transaction.NewSnapshotAuthorityIssuer().Issue(
		snapshotReadOperationID, map[string]int64{"forms": verifiedGenerations["forms"]},
	)
	if err != nil {
		return transaction.VerifiedOperation{}, errSnapshotOperationAdmission
	}
	return operation, nil
}

func (SnapshotOperationAdmission) String() string   { return authorizationRedaction }
func (SnapshotOperationAdmission) GoString() string { return authorizationRedaction }
func (SnapshotOperationAdmission) LogValue() slog.Value {
	return authorizationLogValue()
}
func (admission SnapshotOperationAdmission) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (SnapshotOperationAdmission) MarshalJSON() ([]byte, error) {
	return rejectAuthorizationJSON()
}
