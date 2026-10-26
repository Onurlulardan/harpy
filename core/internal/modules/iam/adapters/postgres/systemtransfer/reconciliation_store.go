package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type ExportReconciliationStore struct {
	runInTenant func(context.Context, uuid.UUID,
		func(context.Context, *exportReconciliationBound) error) error
}

type exportReconciliationBound struct{ database boundDatabase }

func (store *ExportReconciliationStore) inTenant(ctx context.Context,
	claim app.ExportReconciliationClaim,
	execute func(context.Context, *exportReconciliationBound) error,
) error {
	if store == nil || store.runInTenant == nil || ctx == nil || ctx.Err() != nil ||
		claim.Validate() != nil || execute == nil {
		return app.ErrExportReconciliationInvalid
	}
	return store.runInTenant(ctx, claim.TenantID, execute)
}

func (bound *exportReconciliationBound) schema(ctx context.Context) (string, error) {
	if bound == nil {
		return "", app.ErrExportUnavailable
	}
	schema, err := bound.database.CurrentSchema(ctx)
	if err != nil {
		return "", app.ErrExportUnavailable
	}
	return schema, nil
}

var _ app.ExportReconciliationStore = (*ExportReconciliationStore)(nil)
