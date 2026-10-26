package authorization

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PolicyPort interface {
	EvaluateQuery(
		context.Context,
		transaction.TenantScope,
		PolicyRequest,
	) (PolicyDecision, error)
}

type PolicyRegistration struct {
	Owner string
	Port  PolicyPort
}
