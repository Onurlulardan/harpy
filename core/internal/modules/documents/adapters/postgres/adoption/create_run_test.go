package adoption

import "testing"

func TestCreateOperationContract(t *testing.T) {
	if any((*Store).CreateRun) == nil {
		t.Fatal("adoption create persistence is unavailable")
	}
}
