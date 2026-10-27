package processcode

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

type Code = runtimev1.ExitCode

const (
	Clean               = runtimev1.ExitCodeClean
	InvalidStartup      = runtimev1.ExitCodeInvalidStartup
	DependencyPreflight = runtimev1.ExitCodeDependencyPreflight
	ServeFailure        = runtimev1.ExitCodeServeFailure
	DrainForced         = runtimev1.ExitCodeDrainForced
	DeployMigration     = runtimev1.ExitCodeDeployMigration
	OperatorResumable   = runtimev1.ExitCodeDrainForced
	OperatorTerminal    = runtimev1.ExitCodeDeployMigration
)
