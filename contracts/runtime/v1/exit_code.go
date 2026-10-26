package runtimev1

type ExitCode int

const (
	ExitCodeClean               ExitCode = 0
	ExitCodeInvalidStartup      ExitCode = 2
	ExitCodeDependencyPreflight ExitCode = 3
	ExitCodeServeFailure        ExitCode = 4
	ExitCodeDrainForced         ExitCode = 5
	ExitCodeDeployMigration     ExitCode = 10
)

func (code ExitCode) Valid() bool {
	switch code {
	case ExitCodeClean, ExitCodeInvalidStartup, ExitCodeDependencyPreflight,
		ExitCodeServeFailure, ExitCodeDrainForced, ExitCodeDeployMigration:
		return true
	default:
		return false
	}
}

func (code ExitCode) Int() int {
	return int(code)
}
