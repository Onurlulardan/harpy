package folder

import "testing"

func TestDeleteOperationContract(t *testing.T) {
	if any((*store).Delete) == nil {
		t.Fatal("folder delete persistence is unavailable")
	}
}
