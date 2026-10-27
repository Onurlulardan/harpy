package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type OutboundAttemptStore struct {
	runInTenant func(context.Context, uuid.UUID,
		func(context.Context, *exportAttemptBound) error) error
}

type exportAttemptBound struct{ database boundDatabase }

func (store *OutboundAttemptStore) inExportTenant(ctx context.Context,
	claim app.ExportJobClaim, execute func(context.Context, *exportAttemptBound) error,
) error {
	if store == nil || store.runInTenant == nil || ctx == nil || ctx.Err() != nil ||
		claim.Validate() != nil || execute == nil {
		return app.ErrExportWorkerInvalid
	}
	return store.runInTenant(ctx, claim.TenantID, execute)
}

var _ app.ExportAttemptLoader = (*OutboundAttemptStore)(nil)
