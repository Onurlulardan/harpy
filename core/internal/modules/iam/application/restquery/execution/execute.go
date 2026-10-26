package execution

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (executor *Executor) Execute(
	scope transaction.TenantScope,
	program restexecution.Program,
) error {
	if executor == nil || portIsNil(executor.port) || program.Validate() != nil ||
		transaction.RequireActive(scope, transaction.SnapshotQuery) != nil {
		return ErrContract
	}
	if err := executor.port.Execute(scope, program); err != nil {
		return err
	}
	return nil
}
