package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) startExisting(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	identity ScopeIdentity,
	installation catalogdomain.Installation,
	kind catalogdomain.OperationKind,
	idempotencyKey string,
	operationID uuid.UUID,
) (OperationHandle, error) {
	packageID, err := uuidFromCanonical(installation.PackageID.String())
	if err != nil {
		return OperationHandle{}, err
	}
	return application.start(ctx, scope, store, identity, startInput{
		Kind: kind, OperationID: operationID,
		PackageID: packageID,
		Version:   installation.ActiveVersion.String(), IdempotencyKey: idempotencyKey,
		Installation: installation,
		RequestDocument: struct{ Operation, PackageID string }{
			string(kind), installation.PackageID.String(),
		},
	})
}
