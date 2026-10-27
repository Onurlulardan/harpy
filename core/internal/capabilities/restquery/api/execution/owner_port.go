package execution

import "github.com/Onurlulardan/harpy/core/internal/platform/transaction"

// OwnerPort binds its static owner inside scope and synchronously calls Program.Run.
type OwnerPort interface {
	Run(scope transaction.TenantScope, program Program) error
}
