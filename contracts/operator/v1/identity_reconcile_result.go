package operatorv1

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/google/uuid"
)

const IdentityReconcileResultContract = "operator-identity-reconcile-result.v1"

var resultDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type IdentityReconcileResult struct {
	ContractVersion    string `json:"contract_version"`
	Status             string `json:"status"`
	RestoreOperationID string `json:"restore_operation_id"`
	EnvironmentID      string `json:"environment_id"`
	TenantID           string `json:"tenant_id"`
	AdmissionState     string `json:"admission_state"`
	InvalidationSHA256 string `json:"invalidation_sha256"`
	CustodyGeneration  int64  `json:"custody_generation"`
	CustodySHA256      string `json:"custody_result_sha256"`
}

func (result IdentityReconcileResult) Validate() error {
	validUUID := func(value string) bool {
		parsed, err := uuid.Parse(value)
		return err == nil && parsed != uuid.Nil && parsed.String() == value
	}
	if result.ContractVersion != IdentityReconcileResultContract ||
		result.Status != "completed" && result.Status != "completed_noop" ||
		!validUUID(result.RestoreOperationID) || !validUUID(result.EnvironmentID) ||
		!validUUID(result.TenantID) || result.AdmissionState != "identity_reactivation_required" ||
		!resultDigestPattern.MatchString(result.InvalidationSHA256) ||
		result.CustodyGeneration < 1 || !resultDigestPattern.MatchString(result.CustodySHA256) {
		return fmt.Errorf("operator identity reconcile result is invalid")
	}
	return nil
}

func WriteIdentityReconcileResult(writer io.Writer, result IdentityReconcileResult) error {
	if writer == nil || result.Validate() != nil {
		return fmt.Errorf("operator identity reconcile result is invalid")
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}

func ReadIdentityReconcileResult(reader io.Reader) (IdentityReconcileResult, error) {
	if reader == nil {
		return IdentityReconcileResult{}, fmt.Errorf("operator identity reconcile result is invalid")
	}
	decoder := json.NewDecoder(io.LimitReader(reader, 64*1024))
	decoder.DisallowUnknownFields()
	var result IdentityReconcileResult
	if decoder.Decode(&result) != nil || result.Validate() != nil {
		return IdentityReconcileResult{}, fmt.Errorf("operator identity reconcile result is invalid")
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return IdentityReconcileResult{}, fmt.Errorf("operator identity reconcile result is invalid")
	}
	return result, nil
}
