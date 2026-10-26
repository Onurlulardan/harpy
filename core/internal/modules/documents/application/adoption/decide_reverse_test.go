package adoption

import "testing"

func TestDecideOperationContract(t *testing.T) {
	if any((*Application).DecideReverse) == nil {
		t.Fatal("reverse decision operation is unavailable")
	}
}
