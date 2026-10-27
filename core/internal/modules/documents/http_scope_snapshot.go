package documents

import (
	"context"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) snapshot(ctx context.Context, access documentshttp.Access,
	operation string, action func(context.Context, platformtx.TenantScope) error,
) error {
	_, generation, err := runtime.resolveDocuments(ctx, access, operation)
	if err != nil {
		return err
	}
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return documentshttp.ErrUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"documents": generation})
	if err != nil {
		return documentshttp.ErrUnavailable
	}
	return runtime.dependencies.Scopes.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: verified, Mode: platformtx.SnapshotQuery,
	}, action)
}
