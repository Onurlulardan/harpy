package identity

const MaximumBulkDeleteItems = 100

type BulkDeleteReceiptCommand struct {
	OperationID   IdentityID
	ActorID       IdentityID
	ResourceKind  string
	RequestSHA256 [32]byte
	DeletedCount  int
}

func (command BulkDeleteReceiptCommand) Validate() error {
	if command.OperationID == (IdentityID{}) || command.ActorID == (IdentityID{}) ||
		(command.ResourceKind != "user" && command.ResourceKind != "role" &&
			command.ResourceKind != "organization") ||
		command.RequestSHA256 == [32]byte{} || command.DeletedCount < 1 ||
		command.DeletedCount > MaximumBulkDeleteItems {
		return ErrInvalidCommand
	}
	return nil
}
