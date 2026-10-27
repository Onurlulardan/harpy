package systemtransfer

import (
	"context"

	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *AttemptStore) LoadAttempt(ctx context.Context,
	claim systemtransferapp.AttemptClaim,
) (systemtransferapp.AttemptState, error) {
	var state systemtransferapp.AttemptState
	err := store.inTenant(ctx, claim, func(scoped context.Context,
		bound *attemptBound) error {
		schema, err := bound.database.CurrentSchema(scoped)
		if err != nil {
			return systemtransferapp.ErrWorkerUnavailable
		}
		var raw rawAttempt
		err = bound.database.QueryRow(scoped, claimImportWorkerSQL,
			schema, claim.TenantID, claim.ImportID, claim.JobID, claim.AttemptID,
			claim.LeaseOwnerID, claim.LeaseGeneration,
		).Scan(raw.targets()...)
		if err != nil {
			return mapWorkerError(err)
		}
		state, err = raw.decode(claim)
		return err
	})
	return state, err
}
