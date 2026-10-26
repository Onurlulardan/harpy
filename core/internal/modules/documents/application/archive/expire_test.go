package archive

import "testing"

func TestExpireOperationContract(t *testing.T) {
	if any((*ExpiryCommand).Expire) == nil {
		t.Fatal("archive expiry operation is unavailable")
	}
}
