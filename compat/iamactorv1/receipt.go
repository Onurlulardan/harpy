package iamactorv1

import "time"

type Receipt struct{ consumedAt time.Time }

func newReceipt(consumedAt time.Time) (Receipt, error) {
	if consumedAt.IsZero() {
		return Receipt{}, newError(CodeUnavailable)
	}
	return Receipt{consumedAt: consumedAt.UTC()}, nil
}

func (receipt Receipt) ConsumedAt() time.Time { return receipt.consumedAt }

func (receipt Receipt) Valid() bool {
	return !receipt.consumedAt.IsZero() && receipt.consumedAt.Location() == time.UTC
}
