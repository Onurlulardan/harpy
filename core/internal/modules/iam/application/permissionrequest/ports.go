package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Create(context.Context, transaction.TenantScope, CreateCommand) (GroupRef, error)
	Read(context.Context, transaction.TenantScope, ReadQuery) (Group, error)
	List(context.Context, transaction.TenantScope, ListQuery) (Page, error)
	Review(context.Context, transaction.TenantScope, ReviewCommand) (GroupRef, error)
	Cancel(context.Context, transaction.TenantScope, CancelCommand) (GroupRef, error)
	Expire(context.Context, transaction.TenantScope, ExpireCommand) (GroupRef, error)
	Decide(context.Context, transaction.TenantScope, DecideCommand) (DecisionResult, error)
}

type DecisionAuditStore interface {
	WriteDecisionAudit(
		context.Context,
		transaction.TenantScope,
		DecisionRecord,
	) error
	WriteTransitionAudit(
		context.Context,
		transaction.TenantScope,
		TransitionRecord,
	) (int64, error)
}

type DecisionEffects interface {
	RecordPermissionRequestDecision(
		context.Context,
		transaction.TenantScope,
		DecisionRecord,
	) error
	RecordPermissionRequestTransition(
		context.Context,
		transaction.TenantScope,
		TransitionRecord,
	) error
}
