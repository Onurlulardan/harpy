package lifecycle

import (
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func shouldFailInstalling(
	operation catalogdomain.Operation,
	state catalogdomain.InstallationState,
) bool {
	if state != catalogdomain.InstallationInstalling {
		return false
	}
	if operation.Kind != catalogdomain.OperationInstall &&
		operation.Kind != catalogdomain.OperationRetry {
		return false
	}
	return operation.State == catalogdomain.OperationBlocked ||
		operation.State == catalogdomain.OperationFailed
}
