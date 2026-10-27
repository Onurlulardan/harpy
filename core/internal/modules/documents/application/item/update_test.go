package item

import "testing"

func TestUpdateOperationContract(t *testing.T) {
	if any((*Application).Update) == nil {
		t.Fatal("document update operation is unavailable")
	}
}
