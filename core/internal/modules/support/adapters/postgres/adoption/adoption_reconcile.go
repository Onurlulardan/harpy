package adoption

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func (store *Store) AdoptionReconcile(ctx context.Context, command application.LeaseCommand) (
	application.Run, error,
) {
	return store.runCommand(ctx, reconcileSQL, command.OperationID, command.RunID,
		command.LeaseOwner, command.LeaseGeneration, command.ObservedAt)
}
