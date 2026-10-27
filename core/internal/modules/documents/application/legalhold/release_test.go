package legalhold

import "testing"

func TestReleaseOperationContract(t *testing.T) {
	if any((*Application).Release) == nil {
		t.Fatal("legal hold release operation is unavailable")
	}
}
