package item

import "testing"

func TestListOperationContract(t *testing.T) {
	if any((*store).List) == nil {
		t.Fatal("document list persistence is unavailable")
	}
}
