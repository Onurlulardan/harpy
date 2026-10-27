package link

import "testing"

func TestListOperationContract(t *testing.T) {
	if any((*Application).ListReferenceBlockers) == nil {
		t.Fatal("reference blocker list operation is unavailable")
	}
}
