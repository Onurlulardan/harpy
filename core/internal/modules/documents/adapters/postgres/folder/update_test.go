package folder

import "testing"

func TestUpdateOperationContract(t *testing.T) {
	if any((*store).Update) == nil {
		t.Fatal("folder update persistence is unavailable")
	}
}
