package providerapp

type OrderingMode string

const (
	OrderingContiguous             OrderingMode = "contiguous"
	OrderingRequiresReconciliation OrderingMode = "requires_reconciliation"
)

type EventOrdering struct {
	mode     OrderingMode
	sequence int64
	previous int64
}

func ContiguousOrdering(sequence int64, previous int64) (EventOrdering, error) {
	if sequence < 1 || previous < 0 || previous != sequence-1 {
		return EventOrdering{}, ErrNormalizedEvent
	}
	return EventOrdering{
		mode: OrderingContiguous, sequence: sequence, previous: previous,
	}, nil
}

func ReconciliationOrdering() EventOrdering {
	return EventOrdering{mode: OrderingRequiresReconciliation}
}

func (ordering EventOrdering) Mode() OrderingMode { return ordering.mode }
func (ordering EventOrdering) Sequence() int64    { return ordering.sequence }
func (ordering EventOrdering) Previous() int64    { return ordering.previous }

func (ordering EventOrdering) valid() bool {
	if ordering.mode == OrderingRequiresReconciliation {
		return ordering.sequence == 0 && ordering.previous == 0
	}
	validated, err := ContiguousOrdering(ordering.sequence, ordering.previous)
	return err == nil && validated == ordering
}
