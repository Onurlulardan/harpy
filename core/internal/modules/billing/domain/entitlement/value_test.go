package entitlement

import "testing"

func TestValueIsTypedAndHasNoNegativeSentinel(t *testing.T) {
	boolean := BooleanValue(false)
	if value, ok := boolean.Boolean(); !ok || value {
		t.Fatalf("boolean projection = %t, %t", value, ok)
	}
	if _, ok := boolean.Integer(); ok || boolean.Unlimited() {
		t.Fatal("boolean projected as integer")
	}
	integer, err := IntegerValue(0)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := integer.Integer(); !ok || value != 0 {
		t.Fatalf("integer projection = %d, %t", value, ok)
	}
	if _, err := IntegerValue(-1); err == nil {
		t.Fatal("negative unlimited sentinel was accepted")
	}
	if !UnlimitedValue().Unlimited() {
		t.Fatal("typed unlimited value was not projected")
	}
}
