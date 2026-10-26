package campaign

import "time"

type Discount struct {
	id       Identifier
	priority int32
	active   bool
	start    time.Time
	end      *time.Time
}

func NewDiscount(
	id Identifier,
	priority int32,
	active bool,
	start time.Time,
	end *time.Time,
) (Discount, error) {
	canonicalStart, err := canonicalInstant(start)
	if err != nil || !validIdentifier(id) {
		return Discount{}, ErrDiscount
	}
	canonicalFinish, err := canonicalEnd(end, canonicalStart)
	if err != nil {
		return Discount{}, ErrDiscount
	}
	return Discount{id: id, priority: priority, active: active,
		start: canonicalStart, end: canonicalFinish}, nil
}

func (discount Discount) Identifier() Identifier { return discount.id }
func (discount Discount) Priority() int32        { return discount.priority }

func (discount Discount) eligible(instant time.Time) bool {
	return discount.active && contains(discount.start, discount.end, instant)
}
