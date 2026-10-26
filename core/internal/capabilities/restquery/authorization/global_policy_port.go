package authorization

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type GlobalPolicyPort interface {
	EvaluateGlobalQuery(
		context.Context,
		transaction.GlobalScope,
		GlobalPolicyRequest,
	) (PolicyDecision, error)
}
