package upload

import "testing"

func TestPersistForCopyContract(t *testing.T) {
	if any((*Application).PersistForCopy) == nil {
		t.Fatal("Documents copy upload persistence is unavailable")
	}
}
