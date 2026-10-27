package folder

import "testing"

func TestGetOperationContract(t *testing.T) {
	if any((*store).Get) == nil {
		t.Fatal("folder get persistence is unavailable")
	}
}
