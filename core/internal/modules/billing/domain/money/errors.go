package money

import "errors"

var (
	ErrInvalidMoney               = errors.New("money value is invalid")
	ErrUnsupportedRegistryVersion = errors.New("currency registry version is unsupported")
	ErrUnknownCurrency            = errors.New("currency code is unknown")
	ErrMinorUnitNotApplicable     = errors.New("currency has no applicable minor unit")
	ErrAmountOutOfRange           = errors.New("minor units exceed fc_money_v1 bounds")
	ErrCurrencyMismatch           = errors.New("money currencies do not match")
	ErrNegativeQuantity           = errors.New("quantity cannot be negative")
	ErrNegativeGross              = errors.New("gross amount cannot be negative")
	ErrInvalidBasisPoints         = errors.New("basis points must be between 0 and 10000")
	ErrInvalidAllocation          = errors.New("largest-remainder allocation is invalid")
)
