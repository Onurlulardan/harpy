package providerapp

type CommandKind string

const (
	CommandCheckout  CommandKind = "checkout"
	CommandChange    CommandKind = "change"
	CommandCancel    CommandKind = "cancel"
	CommandRefund    CommandKind = "refund"
	CommandReconcile CommandKind = "reconcile"
)

func (kind CommandKind) Valid() bool {
	switch kind {
	case CommandCheckout, CommandChange, CommandCancel,
		CommandRefund, CommandReconcile:
		return true
	default:
		return false
	}
}
