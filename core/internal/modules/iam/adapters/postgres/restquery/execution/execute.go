package executionpostgres

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func (adapter *adapter) Execute(
	scope tenantScope,
	program restexecution.Program,
) error {
	if adapter == nil || program.Validate() != nil || !activeSnapshot(scope) {
		return catalog.ErrContract
	}
	bound, err := adapter.bind(scope)
	if err != nil {
		return catalog.ErrContract
	}
	return program.Run(&driver{scope: scope, database: bound.database})
}
