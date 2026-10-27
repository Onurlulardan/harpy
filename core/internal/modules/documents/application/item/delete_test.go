package item

import "testing"

func TestDeleteOperationContract(t *testing.T) {
	if any((*Application).Delete) == nil {
		t.Fatal("document delete operation is unavailable")
	}
}
