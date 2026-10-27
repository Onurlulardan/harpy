package adoption

import "testing"

func TestAcceptOperationContract(t *testing.T) {
	if any((*Store).Accept) == nil {
		t.Fatal("adoption accept persistence is unavailable")
	}
}
