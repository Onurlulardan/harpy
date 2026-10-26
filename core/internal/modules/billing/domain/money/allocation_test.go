package money

import (
	"errors"
	"reflect"
	"testing"
)

func TestAllocateLargestRemainderUsesLineIDTieBreak(t *testing.T) {
	total := mustMoney(t, 2, "USD")
	lines := []AllocationLine{
		{LineID: testLineID(3), Weight: mustMoney(t, 1, "USD")},
		{LineID: testLineID(1), Weight: mustMoney(t, 1, "USD")},
		{LineID: testLineID(2), Weight: mustMoney(t, 1, "USD")},
	}
	allocations, err := AllocateLargestRemainder(total, lines)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := allocationMinorUnits(allocations), []int64{1, 1, 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("allocations = %v, want %v", got, want)
	}
}

func TestAllocateLargestRemainderIsInputOrderIndependent(t *testing.T) {
	total := mustMoney(t, 10, "USD")
	first := []AllocationLine{
		{LineID: testLineID(1), Weight: mustMoney(t, 1, "USD")},
		{LineID: testLineID(2), Weight: mustMoney(t, 2, "USD")},
	}
	second := []AllocationLine{first[1], first[0]}
	left, err := AllocateLargestRemainder(total, first)
	if err != nil {
		t.Fatal(err)
	}
	right, err := AllocateLargestRemainder(total, second)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) ||
		!reflect.DeepEqual(allocationMinorUnits(left), []int64{3, 7}) {
		t.Fatalf("allocations differ: left=%v right=%v", left, right)
	}
}

func TestAllocateLargestRemainderFailsClosed(t *testing.T) {
	total := mustMoney(t, 1, "USD")
	valid := AllocationLine{LineID: testLineID(1), Weight: mustMoney(t, 1, "USD")}
	tests := []struct {
		name  string
		lines []AllocationLine
		want  error
	}{
		{name: "empty", want: ErrInvalidAllocation},
		{name: "zero ID", lines: []AllocationLine{{Weight: valid.Weight}}, want: ErrInvalidAllocation},
		{name: "duplicate ID", lines: []AllocationLine{valid, valid}, want: ErrInvalidAllocation},
		{name: "negative weight", lines: []AllocationLine{{LineID: testLineID(1), Weight: mustMoney(t, -1, "USD")}}, want: ErrInvalidAllocation},
		{name: "zero weight", lines: []AllocationLine{{LineID: testLineID(1), Weight: mustMoney(t, 0, "USD")}}, want: ErrInvalidAllocation},
		{name: "currency mismatch", lines: []AllocationLine{{LineID: testLineID(1), Weight: mustMoney(t, 1, "EUR")}}, want: ErrCurrencyMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := AllocateLargestRemainder(total, test.lines)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestAllocateZeroAcrossZeroWeights(t *testing.T) {
	lines := []AllocationLine{
		{LineID: testLineID(2), Weight: mustMoney(t, 0, "JPY")},
		{LineID: testLineID(1), Weight: mustMoney(t, 0, "JPY")},
	}
	allocations, err := AllocateLargestRemainder(mustMoney(t, 0, "JPY"), lines)
	if err != nil || !reflect.DeepEqual(allocationMinorUnits(allocations), []int64{0, 0}) {
		t.Fatalf("zero allocations = %v, error = %v", allocations, err)
	}
}

func testLineID(last byte) [16]byte {
	var id [16]byte
	id[len(id)-1] = last
	return id
}

func allocationMinorUnits(allocations []Allocation) []int64 {
	result := make([]int64, 0, len(allocations))
	for _, allocation := range allocations {
		result = append(result, allocation.Amount.MinorUnits())
	}
	return result
}
