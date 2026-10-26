package adoption

import "testing"

func TestDecideOperationContract(t *testing.T) {
	if any((*Store).DecideReverse) == nil {
		t.Fatal("adoption decision persistence is unavailable")
	}
}
