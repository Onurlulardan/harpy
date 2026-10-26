package postrestore

import "github.com/google/uuid"

func (command ReconcileCommand) valid() bool {
	return command.RestoreOperationID != uuid.Nil &&
		command.EnvironmentID != uuid.Nil &&
		command.RecoverySetID != uuid.Nil &&
		command.TenantID != uuid.Nil &&
		command.OwnerUserID != uuid.Nil &&
		command.AnchorID != uuid.Nil &&
		command.NewDatabaseSystemID != uuid.Nil &&
		command.SecurityEpoch != [32]byte{} &&
		command.RestoreReceiptSHA256 != [32]byte{} &&
		command.ProtectedHeadSHA256 != [32]byte{} &&
		command.CustodyResultSHA256 != [32]byte{} &&
		command.CustodyGeneration > 0 &&
		len(command.Receipt) > 1 &&
		len(command.Receipt) <= 64*1024 &&
		!command.OccurredAt.IsZero()
}
