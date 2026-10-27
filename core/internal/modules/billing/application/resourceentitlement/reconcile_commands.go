package resourceentitlement

type ReconcileCommands struct {
	store ReconciliationStore
	usage Usage
	clock Clock
}

func NewReconcileCommands(
	store ReconciliationStore,
	usage Usage,
	clock Clock,
) (*ReconcileCommands, error) {
	if store == nil || usage == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &ReconcileCommands{store: store, usage: usage, clock: clock}, nil
}
