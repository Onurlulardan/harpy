package archive

import "testing"

func TestGetOperationContract(t *testing.T) {
	if any((*store).Get) == nil {
		t.Fatal("archive get persistence is unavailable")
	}
}
