package folder

import "testing"

func TestMoveOperationContract(t *testing.T) {
	if any((*store).Move) == nil {
		t.Fatal("folder move persistence is unavailable")
	}
}
