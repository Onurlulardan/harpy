package adoption

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func (store *Store) AdoptionRecordAdvance(ctx context.Context, command application.ApplyPageCommand) (
	application.Run, error,
) {
	return store.runCommand(ctx, advanceSQL, command.OperationID, command.RunID,
		command.LeaseOwner, command.LeaseGeneration, command.PageNumber,
		nullableUUID(command.CursorBefore), command.SourcePageSHA256[:], command.ObservedAt)
}
