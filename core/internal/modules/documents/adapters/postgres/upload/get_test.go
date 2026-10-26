package upload

import "testing"

func TestGetOperationContract(t *testing.T) {
	if any((*store).Get) == nil {
		t.Fatal("upload get persistence is unavailable")
	}
}
