package operator

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/operatorresult"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

type failureResult struct {
	Success   bool   `json:"success"`
	Stage     string `json:"stage"`
	ExitCode  int    `json:"exit_code"`
	Cause     string `json:"cause,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Detail    string `json:"detail,omitempty"`
	SQLState  string `json:"sqlstate,omitempty"`
}

func writeFailure(writer io.Writer, stage string, code processcode.Code) {
	writeFailureCause(writer, stage, code, "")
}

func writeFailureCause(writer io.Writer, stage string, code processcode.Code, cause string) {
	writeFailureRecord(writer, failureResult{
		Success: false, Stage: stage, ExitCode: code.Int(), Cause: cause,
	})
}

func writeInitializeFailure(
	writer io.Writer, stage string, code processcode.Code, err error,
) {
	if err == nil {
		return
	}
	record := failureResult{
		Success: false, Stage: stage, ExitCode: code.Int(), Cause: err.Error(),
	}
	if diagnostic := environmentoperator.Diagnose(err); diagnostic.Stage != "" {
		record.Stage = diagnostic.Stage
		record.ErrorCode = diagnostic.ErrorCode
		record.Detail = diagnostic.Detail
		record.SQLState = diagnostic.SQLState
	}
	writeFailureRecord(writer, record)
}

func writeFailureRecord(writer io.Writer, record failureResult) {
	if writer == nil {
		return
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(record)
}

func writeSuccess(
	writer io.Writer,
	result environmentoperator.InitializeResult,
) error {
	return operatorresult.WriteInitialize(writer,
		result.Outcome == environmentoperator.OutcomeNoop,
		result.OperationID, result.TenantID)
}

func classifyInitializeError(err error) processcode.Code {
	switch {
	case errors.Is(err, platformpostgres.ErrTenantSchemaInvalid),
		errors.Is(err, environmentoperator.ErrInvalid):
		return processcode.InvalidStartup
	case errors.Is(err, environmentoperator.ErrDependencyMismatch):
		return processcode.DependencyPreflight
	case errors.Is(err, environmentoperator.ErrFailed),
		errors.Is(err, environmentoperator.ErrFenced):
		return processcode.OperatorTerminal
	default:
		return processcode.OperatorResumable
	}
}
