package adoption

import "testing"

func TestRecordOperationContract(t *testing.T) {
	if any((*Store).RecordHit) == nil {
		t.Fatal("compatibility hit persistence is unavailable")
	}
}
