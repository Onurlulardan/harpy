package operator

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/operatorresult"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

type failureResult struct {
	Success  bool   `json:"success"`
	Stage    string `json:"stage"`
	ExitCode int    `json:"exit_code"`
}

func writeFailure(writer io.Writer, stage string, code processcode.Code) {
	if writer == nil {
		return
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(failureResult{Success: false, Stage: stage, ExitCode: code.Int()})
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
	case errors.Is(err, environmentoperator.ErrInvalid):
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
