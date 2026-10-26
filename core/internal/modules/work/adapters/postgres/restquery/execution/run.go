package executionpostgres

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
)

func (owner *adapter) Run(scope tenantScope, program restexecution.Program) error {
	if owner == nil || program.Validate() != nil {
		return workquery.ErrUnavailable
	}
	bound, err := owner.bind(scope)
	if err != nil {
		return workquery.ErrUnavailable
	}
	return program.Run(&driver{scope: scope, database: bound.database})
}
