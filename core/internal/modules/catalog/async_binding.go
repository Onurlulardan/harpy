package catalog

import (
	"context"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *AsyncPort) bindAsync(ctx context.Context, scope platformtx.TenantScope) (
	asyncpostgres.Writer, platformtx.OwnerAuthority, error,
) {
	if port == nil || ctx == nil || ctx.Err() != nil || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.NetworkAllowed(ctx) || platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, platformtx.OwnerAuthority{}, asyncUnavailable()
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "catalog")
	if err != nil || !authority.OwnsScope(scope) {
		return nil, platformtx.OwnerAuthority{}, asyncUnavailable()
	}
	writer, err := asyncpostgres.Bind(scope, authority, port.catalogs)
	if err != nil || writer == nil {
		return nil, platformtx.OwnerAuthority{}, asyncUnavailable()
	}
	return writer, authority, nil
}
