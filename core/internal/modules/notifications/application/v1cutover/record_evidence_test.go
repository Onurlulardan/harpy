package v1cutover

import "testing"

func TestRecordEvidenceRejectsIncompleteInventory(t *testing.T) {
	application, _ := NewApplication(cutoverBinderFake{&cutoverStoreFake{}})
	if _, err := application.RecordEvidence(t.Context(), cutoverTestScope(), Inventory{}); err != ErrInvalid {
		t.Fatalf("RecordEvidence() error = %v", err)
	}
}
