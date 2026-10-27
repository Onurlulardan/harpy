package upload

import "testing"

func TestCreateOperationContract(t *testing.T) {
	if any((*store).Create) == nil {
		t.Fatal("upload create persistence is unavailable")
	}
}
