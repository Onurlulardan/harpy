package transaction

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

const providerBindingOwner = "billing"

type ProviderTargetResolver interface {
	ProviderTarget(context.Context, ProviderLookup) (ProviderTargetEvidence, error)
}

type ProviderResolverBinder struct {
	binder postgres.OwnerBinder[ProviderTargetResolver]
	valid  bool
}

func NewProviderResolverBinder(
	build func(postgres.BoundExecutor) ProviderTargetResolver,
) (ProviderResolverBinder, error) {
	if build == nil {
		return ProviderResolverBinder{}, fmt.Errorf("provider resolver binder is invalid")
	}
	binder, err := postgres.NewOwnerBinder(providerBindingOwner, build)
	if err != nil {
		return ProviderResolverBinder{}, err
	}
	return ProviderResolverBinder{binder: binder, valid: true}, nil
}

func resolveProviderTarget(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	request ProviderScopeRequest,
) (target verifiedProviderTarget, err error, panicValue any) {
	resolver, err := postgres.BindOwner(database, request.Resolver.binder)
	if err != nil || resolver == nil {
		return verifiedProviderTarget{}, ErrProviderResolver, nil
	}
	defer func() { panicValue = recover() }()
	evidence, err := resolver.ProviderTarget(markProviderResolutionContext(ctx), request.Lookup)
	if err != nil {
		return verifiedProviderTarget{}, ErrProviderTarget, nil
	}
	target, err = issueProviderTarget(request.Lookup, evidence)
	if err != nil {
		return verifiedProviderTarget{}, ErrProviderTarget, nil
	}
	return target, nil, nil
}
