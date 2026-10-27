package money

const MaximumBasisPoints = int64(10_000)

type BasisPoints uint16

func NewBasisPoints(value int64) (BasisPoints, error) {
	if value < 0 || value > MaximumBasisPoints {
		return 0, ErrInvalidBasisPoints
	}
	return BasisPoints(value), nil
}

func (value BasisPoints) Value() int64 {
	return int64(value)
}

// ApplyBasisPoints rounds the exact rational result once using half-to-even.
func ApplyBasisPoints(amount Money, rate BasisPoints) (Money, error) {
	if err := amount.validate(); err != nil {
		return Money{}, err
	}
	if rate.Value() > MaximumBasisPoints {
		return Money{}, ErrInvalidBasisPoints
	}
	minorUnits, err := roundedProductRatio(
		amount.minorUnits, rate.Value(), MaximumBasisPoints)
	if err != nil {
		return Money{}, err
	}
	return New(minorUnits, amount.currency, amount.registryVersion)
}
