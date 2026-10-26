package executionpostgres

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
)

func (adapter *adapter) Run(
	scope globalScope,
	program restexecution.GlobalProgram,
) error {
	if adapter == nil || program.Validate() != nil || !activeGlobalSnapshot(scope) {
		return offering.ErrContract
	}
	bound, err := adapter.bind(scope)
	if err != nil {
		return offering.ErrContract
	}
	return program.Run(&driver{scope: scope, database: bound.database})
}
