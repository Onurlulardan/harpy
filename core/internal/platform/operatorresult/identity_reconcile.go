package operatorresult

import (
	"encoding/hex"
	"io"

	operatorv1 "github.com/Onurlulardan/harpy/contracts/operator/v1"
)

type IdentityReconcile struct {
	RestoreOperationID string
	EnvironmentID      string
	TenantID           string
	AdmissionState     string
	InvalidationSHA256 [32]byte
	CustodyGeneration  int64
	CustodySHA256      [32]byte
	Replayed           bool
}

func WriteIdentityReconcile(writer io.Writer, result IdentityReconcile) error {
	status := "completed"
	if result.Replayed {
		status = "completed_noop"
	}
	wire := operatorv1.IdentityReconcileResult{
		ContractVersion:    operatorv1.IdentityReconcileResultContract,
		Status:             status,
		RestoreOperationID: result.RestoreOperationID,
		EnvironmentID:      result.EnvironmentID,
		TenantID:           result.TenantID,
		AdmissionState:     result.AdmissionState,
		InvalidationSHA256: hex.EncodeToString(result.InvalidationSHA256[:]),
		CustodyGeneration:  result.CustodyGeneration,
		CustodySHA256:      hex.EncodeToString(result.CustodySHA256[:]),
	}
	return operatorv1.WriteIdentityReconcileResult(writer, wire)
}
