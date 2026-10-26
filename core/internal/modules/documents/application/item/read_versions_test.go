package item

import "testing"

func TestReadOperationContract(t *testing.T) {
	if any((*Application).ReadVersions) == nil {
		t.Fatal("version read operation is unavailable")
	}
}
