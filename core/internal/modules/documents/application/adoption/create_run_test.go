package adoption

import "testing"

func TestCreateOperationContract(t *testing.T) {
	if any((*Application).CreateRun) == nil {
		t.Fatal("create run operation is unavailable")
	}
}
