package item

import "testing"

func TestMoveOperationContract(t *testing.T) {
	if any((*store).Move) == nil {
		t.Fatal("document move persistence is unavailable")
	}
}
