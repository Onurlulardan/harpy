package subscription

import (
	"regexp"
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type VersionIdentifier [16]byte

type ManualChange struct {
	offeringVersionID VersionIdentifier
	priceVersionID    VersionIdentifier
	effectiveAt       time.Time
	currency          string
}

func ScheduleManualChange(
	state State,
	grantPeriod billingperiod.Period,
	now time.Time,
	currentCurrency string,
	targetCurrency string,
	offeringVersionID VersionIdentifier,
	priceVersionID VersionIdentifier,
) (ManualChange, error) {
	if state != Active && state != Trialing || !grantPeriod.Contains(now) ||
		!currencyPattern.MatchString(currentCurrency) || targetCurrency != currentCurrency ||
		offeringVersionID == (VersionIdentifier{}) || priceVersionID == (VersionIdentifier{}) {
		return ManualChange{}, ErrChange
	}
	return ManualChange{
		offeringVersionID: offeringVersionID, priceVersionID: priceVersionID,
		effectiveAt: grantPeriod.End(), currency: targetCurrency,
	}, nil
}

func (change ManualChange) OfferingVersionID() VersionIdentifier {
	return change.offeringVersionID
}

func (change ManualChange) PriceVersionID() VersionIdentifier { return change.priceVersionID }
func (change ManualChange) EffectiveAt() time.Time            { return change.effectiveAt }
func (change ManualChange) Currency() string                  { return change.currency }
