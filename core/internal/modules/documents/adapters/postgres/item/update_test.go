package item

import "testing"

func TestUpdateOperationContract(t *testing.T) {
	if any((*store).Update) == nil {
		t.Fatal("document update persistence is unavailable")
	}
}
