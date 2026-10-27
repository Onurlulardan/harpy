package archive

import "testing"

func TestCreateOperationContract(t *testing.T) {
	if any((*Application).Create) == nil {
		t.Fatal("archive create operation is unavailable")
	}
}
