package adoption

import "testing"

func TestAcceptOperationContract(t *testing.T) {
	if any((*Application).Accept) == nil {
		t.Fatal("accept operation is unavailable")
	}
}
