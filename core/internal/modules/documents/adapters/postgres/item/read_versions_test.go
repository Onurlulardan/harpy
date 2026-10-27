package item

import "testing"

func TestReadOperationContract(t *testing.T) {
	if any((*store).ReadVersions) == nil {
		t.Fatal("version read persistence is unavailable")
	}
}
