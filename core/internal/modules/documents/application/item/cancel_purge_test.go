package item

import "testing"

func TestCancelOperationContract(t *testing.T) {
	if any((*Application).CancelPurge) == nil {
		t.Fatal("purge cancellation operation is unavailable")
	}
}
