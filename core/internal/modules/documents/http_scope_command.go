package documents

import (
	"context"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) command(ctx context.Context, access documentshttp.Access,
	operation string, action func(context.Context, platformtx.TenantScope) error,
) error {
	if documentsBillingOperation(operation) {
		return runtime.billingCommand(ctx, access, operation, action)
	}
	group, generation, err := runtime.resolveDocuments(ctx, access, operation)
	if err != nil {
		return err
	}
	request, err := platformtx.AdmitCommandScope(access.PrincipalID(), access.TenantID(),
		operation, "documents", group, map[string]int64{"documents": generation})
	if err != nil {
		return documentshttp.ErrUnavailable
	}
	return runtime.dependencies.Scopes.InTenant(ctx, request, action)
}

func documentsBillingOperation(operation string) bool {
	switch operation {
	case "documents.archive.create", "documents.item.copy", "documents.upload.cancel",
		"documents.upload.create", "documents.upload.put":
		return true
	default:
		return false
	}
}
