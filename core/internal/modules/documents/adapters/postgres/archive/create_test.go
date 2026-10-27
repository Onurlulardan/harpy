package archive

import "testing"

func TestCreateOperationContract(t *testing.T) {
	if any((*store).Create) == nil {
		t.Fatal("archive create persistence is unavailable")
	}
}
