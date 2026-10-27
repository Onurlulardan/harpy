package runtime

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CatalogPort interface {
	Catalog(context.Context, transaction.TenantScope, string) (api.CatalogSet, error)
}

type PolicyPort interface {
	EvaluateQuery(
		context.Context, transaction.TenantScope, restauthorization.PolicyRequest,
	) (restauthorization.PolicyDecision, error)
}

type OwnerPort interface {
	Run(transaction.TenantScope, execution.Program) error
}

type GlobalCatalogPort interface {
	Catalog(context.Context, transaction.GlobalScope, string) (api.CatalogSet, error)
}

type GlobalPolicyPort interface {
	EvaluateGlobalQuery(
		context.Context, transaction.GlobalScope, restauthorization.GlobalPolicyRequest,
	) (restauthorization.PolicyDecision, error)
}

type GlobalOwnerPort interface {
	Run(transaction.GlobalScope, execution.GlobalProgram) error
}
