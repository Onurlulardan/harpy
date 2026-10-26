package operator

import (
	"context"
	"encoding/json"
	"io"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

type restoredLicenseResult struct {
	ContractVersion    string `json:"contract_version"`
	Status             string `json:"status"`
	RestoreOperationID string `json:"restore_operation_id"`
	EnvironmentID      string `json:"environment_id"`
	PreviousSequence   int64  `json:"previous_sequence"`
	AcceptedSequence   int64  `json:"accepted_sequence"`
	AuthorityRevision  int64  `json:"authority_revision"`
}

func runRestoredLicenseCommand(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	result, stage, err := reconcileRestoredLicense(ctx)
	if err != nil {
		code := classifyRestoredLicenseError(stage, err)
		writeFailure(stderr, stage, code)
		return code
	}
	status := "completed"
	if result.Replayed {
		status = "completed_noop"
	}
	wire := restoredLicenseResult{
		ContractVersion: "operator-restored-license-result.v1", Status: status,
		RestoreOperationID: result.OperationID.String(),
		EnvironmentID:      result.EnvironmentID.String(),
		PreviousSequence:   result.PreviousSequence,
		AcceptedSequence:   result.AcceptedSequence,
		AuthorityRevision:  result.AuthorityRevision,
	}
	if stdout == nil || json.NewEncoder(stdout).Encode(wire) != nil {
		writeFailure(stderr, "result", processcode.OperatorTerminal)
		return processcode.OperatorTerminal
	}
	return processcode.Clean
}

func classifyRestoredLicenseError(stage string, err error) processcode.Code {
	switch stage {
	case "configuration", "build_contract":
		return processcode.InvalidStartup
	case "database", "composition":
		return processcode.DependencyPreflight
	case "evidence", "verification":
		return processcode.OperatorTerminal
	case "binding", "capability", "reconcile":
		if platformpostgres.ErrorHasSQLState(
			err, "22023", "23505", "42501", "55000", "P0002",
		) {
			return processcode.OperatorTerminal
		}
		return processcode.OperatorResumable
	case "source":
		return processcode.OperatorResumable
	default:
		return processcode.OperatorTerminal
	}
}
