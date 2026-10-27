package execution

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (executor *Executor) Run(
	scope transaction.TenantScope,
	program restexecution.Program,
) error {
	return executor.Execute(scope, program)
}

var _ restexecution.OwnerPort = (*Executor)(nil)
