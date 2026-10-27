package entitlement

import "testing"

func TestDefinitionCombinationMatrix(t *testing.T) {
	key, _ := NewKey("iam.active_users")
	tests := []struct {
		name      string
		value     ValueKind
		meter     MeterKind
		reset     ResetPolicy
		enforce   Enforcement
		unlimited bool
		valid     bool
	}{
		{"boolean none", Boolean, NoMeter, Never, NoEnforcement, false, true},
		{"boolean hard", Boolean, NoMeter, Never, Hard, false, false},
		{"boolean soft", Boolean, NoMeter, Never, Soft, false, false},
		{"boolean metered", Boolean, Counter, Never, Hard, false, false},
		{"boolean unlimited", Boolean, NoMeter, Never, NoEnforcement, true, false},
		{"counter period", Integer, Counter, BillingPeriod, Hard, false, true},
		{"counter month unlimited", Integer, Counter, CalendarMonthUTC, Hard, true, true},
		{"counter never", Integer, Counter, Never, Soft, false, true},
		{"gauge never", Integer, Gauge, Never, Hard, false, true},
		{"gauge reset", Integer, Gauge, CalendarMonthUTC, Hard, false, false},
		{"integer no meter", Integer, NoMeter, Never, Hard, false, false},
		{"unknown enforcement", Integer, Counter, Never, "warn", false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition, err := NewDefinition(
				key, test.value, test.meter, test.reset, test.enforce, test.unlimited,
			)
			if (err == nil) != test.valid {
				t.Fatalf("definition error = %v", err)
			}
			if err == nil && (definition.Key() != key ||
				definition.AllowsUnlimited() != test.unlimited) {
				t.Fatalf("definition projection = %#v", definition)
			}
		})
	}
}
