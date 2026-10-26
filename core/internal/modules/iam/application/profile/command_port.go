package profile

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CommandPort interface {
	SelfPort
	Publish(context.Context, transaction.TenantScope, PublishCommand) (MutationResult, error)
	Backfill(context.Context, transaction.TenantScope, BackfillCommand) (MutationResult, error)
	Retire(context.Context, transaction.TenantScope, RetireCommand) (MutationResult, error)
}

type ReferencePort interface {
	AssertRetirable(
		context.Context,
		transaction.TenantScope,
		RetirementTarget,
	) (ReferenceEvidence, error)
}
