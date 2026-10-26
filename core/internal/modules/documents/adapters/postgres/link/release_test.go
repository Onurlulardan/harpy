package link

import "testing"

func TestReleaseOperationContract(t *testing.T) {
	if any((*store).Release) == nil {
		t.Fatal("reference release persistence is unavailable")
	}
}
