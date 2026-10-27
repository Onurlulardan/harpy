package copy

import "testing"

func TestCopyOperationContract(t *testing.T) {
	if any((*Application).Copy) == nil {
		t.Fatal("document copy operation is unavailable")
	}
}
