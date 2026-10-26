package recoveryanchorpostgres

import "testing"

func TestFinalizeRejectsMissingPool(t *testing.T) {
	if _, err := NewAuthorityFinalizerStore(nil); err == nil {
		t.Fatal("nil authority finalizer pool was accepted")
	}
}
