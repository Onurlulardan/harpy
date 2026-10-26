package systemtransfer

import app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"

func controlAccessMode(access app.ControlScopeAccess,
	scope tenantScope,
) (scopeMode, bool) {
	if access == app.ControlCommandAccess && scope.Mode() == commandScopeMode {
		return commandScopeMode, true
	}
	if access == app.ControlReadAccess && (scope.Mode() == commandScopeMode ||
		scope.Mode() == snapshotQueryScopeMode) {
		return scope.Mode(), true
	}
	return "", false
}
