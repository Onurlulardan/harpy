package profilebackfill

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type RecoveryStore interface {
	StartRecovery(
		context.Context,
		platformtx.TenantScope,
		PreparedRecoveryPlan,
	) (RecoveryResult, error)
}
