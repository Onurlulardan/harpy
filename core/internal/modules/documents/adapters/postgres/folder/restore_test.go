package folder

import "testing"

func TestRestoreOperationContract(t *testing.T) {
	if any((*store).Restore) == nil {
		t.Fatal("folder restore persistence is unavailable")
	}
}
