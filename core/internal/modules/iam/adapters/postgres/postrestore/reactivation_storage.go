package postrestore

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

func (store *Store) PersistReactivation(
	ctx context.Context, command application.CompleteCommand,
) error {
	return store.tenant(ctx, transactionCommand, command.TenantID,
		func(database database) error {
			var applied bool
			return database.QueryRow(ctx, completeSQL,
				command.OperationID, command.EnvironmentID, command.RestoreOperationID,
				command.TenantID, command.OwnerUserID, command.AnchorID,
				command.AssertionSHA256[:], command.PasswordSHA256[:],
				command.ReceiptSHA256[:], command.SignCount,
				command.Password.Salt[:], command.Password.Tag[:],
				int32(command.Password.MemoryKiB), int16(command.Password.Iterations),
				int16(command.Password.Parallelism), command.Identifier.ID,
				command.Identifier.Revision, command.DirectoryKeyID,
				command.DirectoryHMAC[:], command.OccurredAt,
			).Scan(&applied)
		})
}
