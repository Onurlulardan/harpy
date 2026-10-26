package adoption

import "testing"

func TestRecordOperationContract(t *testing.T) {
	if any((*Application).RecordCompatibilityHit) == nil {
		t.Fatal("compatibility hit operation is unavailable")
	}
}
