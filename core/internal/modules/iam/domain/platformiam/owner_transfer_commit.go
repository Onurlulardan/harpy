package platformiam

type OwnerTransferCommit struct {
	Evidence
	OutboxEventID ID
	EventSHA256   [32]byte
}

func (command OwnerTransferCommit) Validate() error {
	if command.Evidence.Validate() != nil || !command.OutboxEventID.Valid() ||
		command.EventSHA256 == [32]byte{} {
		return ErrInvalidCommand
	}
	return nil
}
