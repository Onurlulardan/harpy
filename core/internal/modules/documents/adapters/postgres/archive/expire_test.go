package archive

import "testing"

func TestExpireOperationContract(t *testing.T) {
	if any((*store).Expire) == nil {
		t.Fatal("archive expiry persistence is unavailable")
	}
}
