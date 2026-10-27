package adoption

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func (store *Store) AdoptionClaim(ctx context.Context, command application.ClaimCommand) (
	application.Run, error,
) {
	return store.runCommand(ctx, claimSQL, command.OperationID, command.RunID,
		command.LeaseOwner, command.ExpectedRevision, command.ObservedAt)
}
