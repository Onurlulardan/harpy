package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type AttemptStore struct {
	runInTenant func(context.Context, uuid.UUID,
		func(context.Context, *attemptBound) error) error
}

type attemptBound struct{ database boundDatabase }

func (store *AttemptStore) inTenant(ctx context.Context,
	claim app.AttemptClaim,
	execute func(context.Context, *attemptBound) error,
) error {
	if store == nil || store.runInTenant == nil || ctx == nil || ctx.Err() != nil ||
		claim.Validate() != nil || execute == nil {
		return app.ErrInvalidAttempt
	}
	return store.runInTenant(ctx, claim.TenantID, execute)
}

var _ app.AttemptLoader = (*AttemptStore)(nil)
