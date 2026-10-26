package restquery

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type OwnerPort interface {
	Run(platformtx.TenantScope, restexecution.Program) error
}
