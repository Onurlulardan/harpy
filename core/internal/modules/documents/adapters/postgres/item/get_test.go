package item

import "testing"

func TestGetOperationContract(t *testing.T) {
	if any((*store).Get) == nil {
		t.Fatal("document get persistence is unavailable")
	}
}
