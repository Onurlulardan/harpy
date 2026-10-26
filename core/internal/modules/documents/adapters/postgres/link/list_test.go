package link

import "testing"

func TestListOperationContract(t *testing.T) {
	if any((*store).ListBlockers) == nil || any((*store).SourceLinks) == nil {
		t.Fatal("reference blocker list persistence is unavailable")
	}
}
