package backupreceipt

import "sort"

func validGenerations(values []OwnerGeneration) bool {
	if len(values) == 0 || len(values) > 32 {
		return false
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value.Owner == "" || len(value.Owner) > 64 || value.Generation <= 0 || seen[value.Owner] {
			return false
		}
		seen[value.Owner] = true
	}
	return true
}

func sameGenerations(left, right []OwnerGeneration) bool {
	if !validGenerations(left) || !validGenerations(right) || len(left) != len(right) {
		return false
	}
	leftCopy, rightCopy := append([]OwnerGeneration(nil), left...), append([]OwnerGeneration(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i].Owner < leftCopy[j].Owner })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i].Owner < rightCopy[j].Owner })
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}
