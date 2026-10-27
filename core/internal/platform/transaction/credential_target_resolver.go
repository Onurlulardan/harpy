package transaction

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type CredentialTargetResolver interface {
	CredentialTarget(
		context.Context,
		CredentialLookup,
	) (CredentialTargetEvidence, error)
}

type CredentialResolverBinder struct {
	binder postgres.OwnerBinder[CredentialTargetResolver]
	valid  bool
}

func NewCredentialResolverBinder(
	build func(postgres.BoundExecutor) CredentialTargetResolver,
) (CredentialResolverBinder, error) {
	if build == nil {
		return CredentialResolverBinder{}, fmt.Errorf("credential resolver binder is invalid")
	}
	binder, err := postgres.NewOwnerBinder(credentialBindingOwner, build)
	if err != nil {
		return CredentialResolverBinder{}, err
	}
	return CredentialResolverBinder{binder: binder, valid: true}, nil
}

func resolveCredentialTarget(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	request CredentialScopeRequest,
) (target VerifiedCredentialTarget, err error, panicValue any) {
	resolver, err := postgres.BindOwner(database, request.Resolver.binder)
	if err != nil || resolver == nil {
		return VerifiedCredentialTarget{}, ErrCredentialResolver, nil
	}
	defer func() { panicValue = recover() }()
	evidence, err := resolver.CredentialTarget(
		markCredentialResolutionContext(ctx), request.Lookup,
	)
	if err != nil {
		return VerifiedCredentialTarget{}, ErrCredentialTarget, nil
	}
	target, err = issueCredentialTarget(evidence)
	if err != nil || target.operationID != request.Lookup.operationID ||
		target.purpose != request.Lookup.purpose ||
		target.locatorKind != request.Lookup.locatorKind ||
		!target.matchesLookup(request.Lookup) ||
		(request.Lookup.tenantID != [16]byte{} &&
			target.tenantID != request.Lookup.tenantID) {
		return VerifiedCredentialTarget{}, ErrCredentialTarget, nil
	}
	return target, nil, nil
}
