package item

import "testing"

func TestRestoreOperationContract(t *testing.T) {
	if any((*store).Restore) == nil {
		t.Fatal("document restore persistence is unavailable")
	}
}
