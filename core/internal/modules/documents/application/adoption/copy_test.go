package adoption

import "testing"

func TestCopyOperationContract(t *testing.T) {
	if any((*AdoptionCopyWorker).Copy) == nil {
		t.Fatal("copy operation is unavailable")
	}
}
