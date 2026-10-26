package actorbridge

import (
	"context"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type TenantCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type AuthorizationPort interface {
	AuthorizeActorBridge(
		context.Context,
		platformtx.TenantScope,
		string,
		string,
		int64,
		int64,
	) (Authorization, error)
}

type ScopedIssueAuthorityPort interface {
	ResolveScopedIssueAuthority(
		context.Context,
		platformtx.TenantScope,
		actorassertionv1.Domain,
	) (Authority, error)
}
