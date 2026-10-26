package link

import "testing"

func TestReleaseOperationContract(t *testing.T) {
	if any((*Application).Release) == nil {
		t.Fatal("reference release operation is unavailable")
	}
}
