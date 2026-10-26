package archive

import "testing"

func TestGetOperationContract(t *testing.T) {
	if any((*Application).Get) == nil {
		t.Fatal("archive get operation is unavailable")
	}
}
