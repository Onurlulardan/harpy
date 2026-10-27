package entitlement

type ValueKind string
type MeterKind string
type ResetPolicy string
type Enforcement string

const (
	Boolean ValueKind = "boolean"
	Integer ValueKind = "integer"

	NoMeter MeterKind = "none"
	Counter MeterKind = "counter"
	Gauge   MeterKind = "gauge"

	BillingPeriod    ResetPolicy = "billing_period"
	CalendarMonthUTC ResetPolicy = "calendar_month_utc"
	Never            ResetPolicy = "never"

	NoEnforcement Enforcement = "none"
	Hard          Enforcement = "hard"
	Soft          Enforcement = "soft"
)

type Definition struct {
	key            Key
	valueKind      ValueKind
	meterKind      MeterKind
	resetPolicy    ResetPolicy
	enforcement    Enforcement
	allowUnlimited bool
}

func NewDefinition(
	key Key,
	valueKind ValueKind,
	meterKind MeterKind,
	resetPolicy ResetPolicy,
	enforcement Enforcement,
	allowUnlimited bool,
) (Definition, error) {
	definition := Definition{
		key: key, valueKind: valueKind, meterKind: meterKind,
		resetPolicy: resetPolicy, enforcement: enforcement,
		allowUnlimited: allowUnlimited,
	}
	if !definition.valid() {
		return Definition{}, ErrDefinition
	}
	return definition, nil
}

func (definition Definition) Key() Key                 { return definition.key }
func (definition Definition) ValueKind() ValueKind     { return definition.valueKind }
func (definition Definition) MeterKind() MeterKind     { return definition.meterKind }
func (definition Definition) ResetPolicy() ResetPolicy { return definition.resetPolicy }
func (definition Definition) Enforcement() Enforcement { return definition.enforcement }
func (definition Definition) AllowsUnlimited() bool    { return definition.allowUnlimited }

func (definition Definition) valid() bool {
	if !definition.key.Valid() {
		return false
	}
	if definition.valueKind == Boolean {
		return definition.meterKind == NoMeter && definition.resetPolicy == Never &&
			definition.enforcement == NoEnforcement && !definition.allowUnlimited
	}
	if definition.valueKind != Integer ||
		(definition.meterKind != Counter && definition.meterKind != Gauge) ||
		(definition.enforcement != Hard && definition.enforcement != Soft) {
		return false
	}
	if definition.meterKind == Gauge {
		return definition.resetPolicy == Never
	}
	return definition.resetPolicy == BillingPeriod ||
		definition.resetPolicy == CalendarMonthUTC || definition.resetPolicy == Never
}
