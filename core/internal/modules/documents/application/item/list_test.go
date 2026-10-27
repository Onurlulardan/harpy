package item

import "testing"

func TestListOperationContract(t *testing.T) {
	if any((*Application).List) == nil {
		t.Fatal("document list operation is unavailable")
	}
}
