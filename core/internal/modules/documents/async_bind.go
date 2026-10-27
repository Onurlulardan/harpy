package documents

import (
	"context"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *APIAsyncPort) bind(ctx context.Context,
	scope platformtx.TenantScope,
) (asyncpostgres.Writer, platformtx.OwnerAuthority, error) {
	if port == nil || ctx == nil || ctx.Err() != nil ||
		!platformtx.ContextOwnsScope(ctx, scope) || platformtx.NetworkAllowed(ctx) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, platformtx.OwnerAuthority{}, ErrDocumentsRuntime
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "documents")
	if err != nil || !authority.OwnsScope(scope) {
		return nil, platformtx.OwnerAuthority{}, ErrDocumentsRuntime
	}
	writer, err := asyncpostgres.Bind(scope, authority, port.catalogs)
	if err != nil || writer == nil {
		return nil, platformtx.OwnerAuthority{}, ErrDocumentsRuntime
	}
	return writer, authority, nil
}
