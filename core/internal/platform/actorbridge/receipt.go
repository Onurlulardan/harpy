package actorbridge

import "time"

type Receipt struct {
	consumedAt time.Time
}

func NewReceipt(consumedAt time.Time) (Receipt, error) {
	if consumedAt.IsZero() {
		return Receipt{}, NewFault(FaultInvalid)
	}
	return Receipt{consumedAt: consumedAt.UTC()}, nil
}

func (receipt Receipt) Validate() error {
	if receipt.consumedAt.IsZero() || receipt.consumedAt.Location() != time.UTC {
		return NewFault(FaultInvalid)
	}
	return nil
}

func (receipt Receipt) ConsumedAt() time.Time {
	if receipt.Validate() != nil {
		return time.Time{}
	}
	return receipt.consumedAt
}
