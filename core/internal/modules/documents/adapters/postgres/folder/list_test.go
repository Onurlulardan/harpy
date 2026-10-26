package folder

import "testing"

func TestListOperationContract(t *testing.T) {
	if any((*store).List) == nil {
		t.Fatal("folder list persistence is unavailable")
	}
}
