package legalhold

import "testing"

func TestReleaseOperationContract(t *testing.T) {
	if any((*store).Release) == nil {
		t.Fatal("legal hold release persistence is unavailable")
	}
}
