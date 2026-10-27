package billingv1adoptionapp

import (
	"math"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

const bytesPerLegacyMegabyte = int64(1_048_576)

func integerValue(value int64) (entitlement.Value, error) {
	if value == -1 {
		return entitlement.UnlimitedValue(), nil
	}
	return entitlement.IntegerValue(value)
}

func storageValue(megabytes int64) (entitlement.Value, error) {
	if megabytes == -1 {
		return entitlement.UnlimitedValue(), nil
	}
	if megabytes < 0 || megabytes > math.MaxInt64/bytesPerLegacyMegabyte {
		return entitlement.Value{}, ErrLimits
	}
	bytes := megabytes * bytesPerLegacyMegabyte
	if bytes > 9_000_000_000_000_000 {
		return entitlement.Value{}, ErrLimits
	}
	return entitlement.IntegerValue(bytes)
}
