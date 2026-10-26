package operator

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/platform/operatorresult"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func runRestoredIdentityCommand(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	result, stage, err := reconcileRestoredIdentity(ctx)
	if err != nil {
		code := classifyIdentityReconcileError(stage, err)
		writeFailure(stderr, stage, code)
		return code
	}
	wire := operatorresult.IdentityReconcile{
		RestoreOperationID: result.OperationID.String(), EnvironmentID: result.EnvironmentID.String(),
		TenantID: result.TenantID.String(), AdmissionState: result.State, Replayed: result.Replayed,
		InvalidationSHA256: result.InvalidationSHA256,
		CustodyGeneration:  result.CustodyGeneration,
		CustodySHA256:      result.CustodySHA256,
	}
	if operatorresult.WriteIdentityReconcile(stdout, wire) != nil {
		writeFailure(stderr, "result", processcode.OperatorTerminal)
		return processcode.OperatorTerminal
	}
	return processcode.Clean
}

func classifyIdentityReconcileError(stage string, err error) processcode.Code {
	switch stage {
	case "configuration", "build_contract":
		return processcode.InvalidStartup
	case "database", "composition":
		return processcode.DependencyPreflight
	case "evidence":
		return processcode.OperatorTerminal
	case "custody":
		return processcode.OperatorResumable
	case "reconcile":
		if platformpostgres.ErrorHasSQLState(err, "22023", "23505", "42501", "P0002") {
			return processcode.OperatorTerminal
		}
		return processcode.OperatorResumable
	default:
		return processcode.OperatorTerminal
	}
}
