package money

import (
	"testing"
	"testing/quick"
)

func TestAddSubtractRoundTripProperty(t *testing.T) {
	property := func(leftRaw, rightRaw int64) bool {
		bound := MaximumAbsoluteMinorUnits / 2
		left := moneyForProperty(leftRaw%(bound+1), "USD")
		right := moneyForProperty(rightRaw%(bound+1), "USD")
		total, err := Add(left, right)
		if err != nil {
			return false
		}
		recovered, err := Subtract(total, right)
		return err == nil && recovered == left
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 2_000}); err != nil {
		t.Fatal(err)
	}
}

func TestBasisPointsMagnitudeProperty(t *testing.T) {
	property := func(amountRaw int64, rateRaw uint16) bool {
		amount := amountRaw % (MaximumAbsoluteMinorUnits + 1)
		if amount < 0 {
			amount = -amount
		}
		rate, err := NewBasisPoints(int64(rateRaw) % (MaximumBasisPoints + 1))
		if err != nil {
			return false
		}
		result, err := ApplyBasisPoints(moneyForProperty(amount, "USD"), rate)
		return err == nil && result.MinorUnits() >= 0 && result.MinorUnits() <= amount
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 2_000}); err != nil {
		t.Fatal(err)
	}
}

func TestLargestRemainderConservesTotalProperty(t *testing.T) {
	property := func(totalRaw, firstRaw, secondRaw uint64) bool {
		total := int64(totalRaw % uint64(MaximumAbsoluteMinorUnits+1))
		first := int64(firstRaw%1_000_000) + 1
		second := int64(secondRaw%1_000_000) + 1
		lines := []AllocationLine{
			{LineID: testLineID(1), Weight: moneyForProperty(first, "USD")},
			{LineID: testLineID(2), Weight: moneyForProperty(second, "USD")},
		}
		allocations, err := AllocateLargestRemainder(moneyForProperty(total, "USD"), lines)
		if err != nil {
			return false
		}
		return allocations[0].Amount.MinorUnits()+allocations[1].Amount.MinorUnits() == total
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 2_000}); err != nil {
		t.Fatal(err)
	}
}

func moneyForProperty(minor int64, currency string) Money {
	value, err := New(minor, currency, RegistryVersion)
	if err != nil {
		panic(err)
	}
	return value
}
