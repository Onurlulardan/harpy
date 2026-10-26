package folder

import "testing"

func TestCreateOperationContract(t *testing.T) {
	if any((*store).Create) == nil {
		t.Fatal("folder create persistence is unavailable")
	}
}
