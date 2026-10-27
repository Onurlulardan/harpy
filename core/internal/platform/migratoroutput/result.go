package migratoroutput

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

type Result struct {
	value runtimev1.MigratorCommandResult
}

func newResult(value runtimev1.MigratorCommandResult) (Result, error) {
	if err := value.Validate(); err != nil {
		return Result{}, err
	}
	return Result{value: value}, nil
}
