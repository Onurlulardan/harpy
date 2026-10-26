package recovery

import "crypto/subtle"

func ValidateInventory(
	items []InventoryItem,
	expected [32]byte,
) ([]InventoryItem, error) {
	normalized, actual, err := normalizeInventory(items)
	if err != nil || subtle.ConstantTimeCompare(actual[:], expected[:]) != 1 {
		return nil, ErrUnavailable
	}
	return normalized, nil
}
