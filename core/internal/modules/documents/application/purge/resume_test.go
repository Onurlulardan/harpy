package purge

import "testing"

func TestResumeOperationContract(t *testing.T) {
	if any((*Command).Resume) == nil {
		t.Fatal("purge resume operation is unavailable")
	}
}
