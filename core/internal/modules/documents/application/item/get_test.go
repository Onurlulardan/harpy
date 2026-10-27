package item

import "testing"

func TestGetOperationContract(t *testing.T) {
	if any((*Application).Get) == nil {
		t.Fatal("document get operation is unavailable")
	}
}
