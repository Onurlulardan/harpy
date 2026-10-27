package money

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
)

type AllocationLine struct {
	LineID [16]byte
	Weight Money
}

type Allocation struct {
	LineID [16]byte
	Amount Money
}

type allocationCandidate struct {
	lineID    [16]byte
	minorUnit int64
	remainder *big.Int
}

// AllocateLargestRemainder distributes a non-negative amount by money weights.
// Results are returned in ascending LineID order, independent of input order.
func AllocateLargestRemainder(
	total Money,
	lines []AllocationLine,
) ([]Allocation, error) {
	if err := total.validate(); err != nil {
		return nil, err
	}
	if total.minorUnits < 0 || len(lines) == 0 {
		return nil, ErrInvalidAllocation
	}
	totalWeight, err := validateAllocationLines(total, lines)
	if err != nil {
		return nil, err
	}
	if totalWeight.Sign() == 0 {
		if total.minorUnits != 0 {
			return nil, ErrInvalidAllocation
		}
		return zeroAllocations(total, lines)
	}
	candidates, allocated, err := floorAllocation(total, lines, totalWeight)
	if err != nil {
		return nil, err
	}
	remaining := new(big.Int).Sub(
		new(big.Int).SetInt64(total.minorUnits), allocated)
	if !remaining.IsInt64() || remaining.Sign() < 0 ||
		remaining.Cmp(big.NewInt(int64(len(candidates)))) > 0 {
		return nil, ErrInvalidAllocation
	}
	sort.Slice(candidates, func(left, right int) bool {
		comparison := candidates[left].remainder.Cmp(candidates[right].remainder)
		if comparison != 0 {
			return comparison > 0
		}
		return bytes.Compare(candidates[left].lineID[:], candidates[right].lineID[:]) < 0
	})
	for index := int64(0); index < remaining.Int64(); index++ {
		next, err := checkedAdd(candidates[index].minorUnit, 1)
		if err != nil {
			return nil, err
		}
		candidates[index].minorUnit = next
	}
	return buildAllocations(total, candidates)
}

func validateAllocationLines(total Money, lines []AllocationLine) (*big.Int, error) {
	totalWeight := new(big.Int)
	seen := make(map[[16]byte]struct{}, len(lines))
	for _, line := range lines {
		if line.LineID == [16]byte{} {
			return nil, fmt.Errorf("%w: zero line ID", ErrInvalidAllocation)
		}
		if _, exists := seen[line.LineID]; exists {
			return nil, fmt.Errorf("%w: duplicate line ID", ErrInvalidAllocation)
		}
		seen[line.LineID] = struct{}{}
		if err := requireSameCurrency(total, line.Weight); err != nil {
			return nil, err
		}
		if line.Weight.minorUnits < 0 {
			return nil, fmt.Errorf("%w: negative weight", ErrInvalidAllocation)
		}
		totalWeight.Add(totalWeight, new(big.Int).SetInt64(line.Weight.minorUnits))
	}
	return totalWeight, nil
}

func floorAllocation(
	total Money,
	lines []AllocationLine,
	totalWeight *big.Int,
) ([]allocationCandidate, *big.Int, error) {
	candidates := make([]allocationCandidate, 0, len(lines))
	allocated := new(big.Int)
	for _, line := range lines {
		numerator := new(big.Int).Mul(
			new(big.Int).SetInt64(total.minorUnits),
			new(big.Int).SetInt64(line.Weight.minorUnits),
		)
		quotient := new(big.Int)
		remainder := new(big.Int)
		quotient.QuoRem(numerator, totalWeight, remainder)
		minorUnit, err := checkedMinor(quotient)
		if err != nil {
			return nil, nil, err
		}
		allocated.Add(allocated, quotient)
		candidates = append(candidates, allocationCandidate{
			lineID: line.LineID, minorUnit: minorUnit, remainder: remainder,
		})
	}
	return candidates, allocated, nil
}

func zeroAllocations(total Money, lines []AllocationLine) ([]Allocation, error) {
	candidates := make([]allocationCandidate, 0, len(lines))
	for _, line := range lines {
		candidates = append(candidates, allocationCandidate{lineID: line.LineID})
	}
	return buildAllocations(total, candidates)
}

func buildAllocations(total Money, candidates []allocationCandidate) ([]Allocation, error) {
	sort.Slice(candidates, func(left, right int) bool {
		return bytes.Compare(candidates[left].lineID[:], candidates[right].lineID[:]) < 0
	})
	result := make([]Allocation, 0, len(candidates))
	for _, candidate := range candidates {
		amount, err := New(candidate.minorUnit, total.currency, total.registryVersion)
		if err != nil {
			return nil, err
		}
		result = append(result, Allocation{LineID: candidate.lineID, Amount: amount})
	}
	return result, nil
}
