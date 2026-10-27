package recovery

import "testing"

func TestValidateInventoryRequiresExactDigest(t *testing.T) {
	item := validInventoryFixture(
		"10000000-0000-4000-8000-000000000001",
	)
	_, digest, err := normalizeInventory([]InventoryItem{item})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ValidateInventory([]InventoryItem{item}, digest); err != nil {
		t.Fatalf("valid inventory rejected: %v", err)
	}
	digest[0] ^= 0xff
	if _, err = ValidateInventory([]InventoryItem{item}, digest); err != ErrUnavailable {
		t.Fatalf("mismatched digest error = %v", err)
	}
}
