package item

import "testing"

func TestMoveOperationContract(t *testing.T) {
	if any((*Application).Move) == nil {
		t.Fatal("document move operation is unavailable")
	}
}
