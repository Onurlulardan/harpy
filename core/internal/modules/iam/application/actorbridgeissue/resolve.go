package actorbridgeissue

import (
	"context"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (resolver *Resolver) Resolve(
	ctx context.Context,
	scope TenantScope,
	domain actorassertionv1.Domain,
) (Evidence, error) {
	if resolver == nil || !resolver.valid || resolver.store == nil || ctx == nil ||
		ctx.Err() != nil || domain.Validate() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) {
		return Evidence{}, ErrUnavailable
	}
	evidence, err := resolver.store.Resolve(ctx, scope, domain)
	if err != nil || evidence.Validate() != nil {
		return Evidence{}, ErrUnavailable
	}
	return evidence, nil
}
