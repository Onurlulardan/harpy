package upload

import "testing"

func TestPrepareForCopyContract(t *testing.T) {
	if any((*Application).PrepareForCopy) == nil {
		t.Fatal("Documents copy upload preparation is unavailable")
	}
}
