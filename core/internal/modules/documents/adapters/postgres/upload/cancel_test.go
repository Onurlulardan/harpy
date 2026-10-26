package upload

import "testing"

func TestCancelOperationContract(t *testing.T) {
	if any((*store).Cancel) == nil {
		t.Fatal("upload cancel persistence is unavailable")
	}
}
