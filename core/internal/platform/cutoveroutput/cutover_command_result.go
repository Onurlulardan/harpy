package cutoveroutput

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

type Result struct {
	value runtimev1.CutoverCommandResult
}

func newResult(value runtimev1.CutoverCommandResult, err error) (Result, error) {
	if err != nil {
		return Result{}, fmt.Errorf("build cutover command result: %w", err)
	}
	if err := value.Validate(); err != nil {
		return Result{}, fmt.Errorf("validate cutover command result: %w", err)
	}
	return Result{value: value}, nil
}

func (result Result) ExitCode() processcode.Code {
	return result.value.ExitCode
}
