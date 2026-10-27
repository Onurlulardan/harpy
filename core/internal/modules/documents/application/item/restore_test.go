package item

import "testing"

func TestRestoreOperationContract(t *testing.T) {
	if any((*Application).Restore) == nil {
		t.Fatal("document restore operation is unavailable")
	}
}
