package v1query

import "slices"

// RetirementKeys returns the canonical lock order for aliases used by this request.
func (result pendingResult) RetirementKeys() ([]RetirementKey, error) {
	if result.Validate() != nil {
		return nil, result.Validate()
	}
	aliasIDs, err := result.AliasIDs()
	if err != nil {
		return nil, err
	}
	keys := make([]RetirementKey, len(aliasIDs))
	for index, aliasID := range aliasIDs {
		key, err := NewRetirementKey(result.manifestID, aliasID)
		if err != nil {
			return nil, err
		}
		keys[index] = key
	}
	slices.SortFunc(keys, func(left, right RetirementKey) int {
		return compareRetirementKeys(left, right)
	})
	return keys, nil
}

func compareRetirementKeys(left, right RetirementKey) int {
	leftValue, leftErr := left.Value()
	rightValue, rightErr := right.Value()
	if leftErr != nil || rightErr != nil {
		return 0
	}
	if leftValue < rightValue {
		return -1
	}
	if leftValue > rightValue {
		return 1
	}
	return 0
}
