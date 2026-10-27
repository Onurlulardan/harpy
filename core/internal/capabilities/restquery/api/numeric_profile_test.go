package api

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

func TestIntegerProfileAcceptsSigned64Bounds(t *testing.T) {
	profile, err := NewIntegerProfile(math.MinInt64, math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	minimum, minimumErr := profile.Minimum()
	maximum, maximumErr := profile.Maximum()
	if minimumErr != nil || maximumErr != nil ||
		minimum != math.MinInt64 || maximum != math.MaxInt64 {
		t.Fatal("integer bounds changed")
	}
	if _, err := NewIntegerProfile(1, 0); err == nil {
		t.Fatal("inverted integer range accepted")
	}
}

func TestDecimalProfileEnforcesPrecisionAndScale(t *testing.T) {
	for _, test := range []struct {
		precision int
		scale     int
		valid     bool
	}{
		{1, 0, true}, {1, 1, true}, {1000, 1000, true},
		{0, 0, false}, {1001, 0, false}, {5, -1, false}, {5, 6, false},
	} {
		profile, err := NewDecimalProfile(test.precision, test.scale)
		if (err == nil) != test.valid {
			t.Fatalf("valid=%v precision=%d scale=%d err=%v",
				test.valid, test.precision, test.scale, err)
		}
		if test.valid {
			precision, precisionErr := profile.Precision()
			scale, scaleErr := profile.Scale()
			if precisionErr != nil || scaleErr != nil ||
				precision != test.precision || scale != test.scale {
				t.Fatal("decimal profile changed")
			}
		}
	}
}

func TestNumericProfilesFailClosedAndRedact(t *testing.T) {
	integer, _ := NewIntegerProfile(-10, 10)
	decimal, _ := NewDecimalProfile(18, 4)
	for _, profile := range []any{integer, decimal} {
		if fmt.Sprint(profile) != queryValueRedaction || fmt.Sprintf("%#v", profile) != queryValueRedaction {
			t.Fatalf("%T formatting leaked", profile)
		}
		if _, err := json.Marshal(profile); err == nil {
			t.Fatalf("%T silently marshaled", profile)
		}
	}
	if err := (IntegerProfile{}).Validate(); err == nil {
		t.Fatal("zero integer profile validated")
	}
	if err := (DecimalProfile{}).Validate(); err == nil {
		t.Fatal("zero decimal profile validated")
	}
	zeroInteger := IntegerProfile{}
	for _, access := range []func() error{
		func() error { _, err := zeroInteger.Minimum(); return err },
		func() error { _, err := zeroInteger.Maximum(); return err },
	} {
		if access() == nil {
			t.Fatal("zero integer accessor succeeded")
		}
	}
	zeroDecimal := DecimalProfile{}
	for _, access := range []func() error{
		func() error { _, err := zeroDecimal.Precision(); return err },
		func() error { _, err := zeroDecimal.Scale(); return err },
	} {
		if access() == nil {
			t.Fatal("zero decimal accessor succeeded")
		}
	}
	integer.minimum = 11
	if err := integer.Validate(); err == nil {
		t.Fatal("tampered integer profile validated")
	}
	if _, err := integer.Maximum(); err == nil {
		t.Fatal("tampered integer accessor succeeded")
	}
	decimal.scale = 19
	if _, err := decimal.Scale(); err == nil {
		t.Fatal("tampered decimal accessor succeeded")
	}
}
