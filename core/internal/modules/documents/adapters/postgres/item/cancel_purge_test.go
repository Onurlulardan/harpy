package item

import "testing"

func TestCancelOperationContract(t *testing.T) {
	if any((*store).CancelPurge) == nil {
		t.Fatal("purge cancellation persistence is unavailable")
	}
}
