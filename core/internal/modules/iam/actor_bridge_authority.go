package iam

import (
	"context"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	actorbridgeissuepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/actorbridgeissue"
	actorbridgeissueapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/actorbridgeissue"
	platformactorbridge "github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type actorBridgeAuthority struct {
	resolver *actorbridgeissueapp.Resolver
	valid    bool
}

func newActorBridgeAuthority() (*actorBridgeAuthority, error) {
	store, err := actorbridgeissuepostgres.NewStore()
	if err != nil {
		return nil, platformactorbridge.NewFault(platformactorbridge.FaultUnavailable)
	}
	resolver, err := actorbridgeissueapp.NewResolver(store)
	if err != nil {
		return nil, platformactorbridge.NewFault(platformactorbridge.FaultUnavailable)
	}
	return &actorBridgeAuthority{resolver: resolver, valid: true}, nil
}

func (authority *actorBridgeAuthority) ResolveScopedIssueAuthority(
	ctx context.Context,
	scope platformtx.TenantScope,
	domain actorassertionv1.Domain,
) (platformactorbridge.Authority, error) {
	if authority == nil || !authority.valid || ctx == nil || ctx.Err() != nil ||
		domain.Validate() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) {
		return platformactorbridge.Authority{},
			platformactorbridge.NewFault(platformactorbridge.FaultInvalid)
	}
	resolved, err := authority.resolver.Resolve(ctx, scope, domain)
	if err != nil {
		return platformactorbridge.Authority{},
			platformactorbridge.NewFault(platformactorbridge.FaultAuthority)
	}
	result, err := platformactorbridge.NewAuthority(
		resolved.IAMGroupGeneration(),
		resolved.IAMGeneration(),
		resolved.TargetGeneration(),
	)
	if err != nil {
		return platformactorbridge.Authority{},
			platformactorbridge.NewFault(platformactorbridge.FaultUnavailable)
	}
	return result, nil
}

var _ platformactorbridge.ScopedIssueAuthorityPort = (*actorBridgeAuthority)(nil)
