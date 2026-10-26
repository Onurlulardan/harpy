package discount

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func validDigest(value [32]byte) bool { return value != [32]byte{} }

func validLimit(limit *int64, used int64) bool {
	return used >= 0 && (limit == nil || *limit > 0 && used <= *limit)
}

func (snapshot EligibilitySnapshot) validFor(request RedeemRequest, now time.Time) bool {
	return validV7(snapshot.DiscountVersionID) && snapshot.Code.Valid() &&
		snapshot.Code == request.Code && snapshot.OfferingVersionID == request.OfferingVersionID &&
		snapshot.PriceVersionID == request.PriceVersionID && validInstant(snapshot.ValidFrom) &&
		validInstant(snapshot.ValidUntil) && snapshot.ValidFrom.Before(snapshot.ValidUntil) &&
		!now.Before(snapshot.ValidFrom) && now.Before(snapshot.ValidUntil) &&
		(snapshot.RequiredCurrency == "" || snapshot.RequiredCurrency == request.Currency) &&
		validLimit(snapshot.GlobalUseLimit, snapshot.GlobalUsed) &&
		validLimit(snapshot.TenantUseLimit, snapshot.TenantUsed) && validDigest(snapshot.PolicySHA256)
}

func exhausted(limit *int64, used int64) bool {
	return limit != nil && used >= *limit
}
