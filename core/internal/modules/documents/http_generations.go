package documents

import (
	"context"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
)

func (runtime *HTTPRuntime) resolveDocuments(ctx context.Context,
	access documentshttp.Access, operation string,
) (int64, int64, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil ||
		!access.ValidFor(access.Permission(), operation) {
		return 0, 0, documentshttp.ErrForbidden
	}
	group, generation, err := runtime.dependencies.Generations.
		ResolveTenantDocumentsGenerations(ctx, access.TenantID())
	if err != nil || group < 1 || generation < 1 {
		return 0, 0, documentshttp.ErrUnavailable
	}
	return group, generation, nil
}
