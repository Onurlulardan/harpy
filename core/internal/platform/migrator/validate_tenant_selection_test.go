package migrator

import "testing"

func TestValidateTenantSelectionRejectsCurrentEmptySelection(t *testing.T) {
	err := validateTenantSelection(nil)
	if err == nil || err.Error() != "tenant migration selection is empty" {
		t.Fatalf("empty current tenant selection error = %v", err)
	}
}
