package adoption

import "testing"

func TestReceiveLegacyArchiveHitOperationExists(t *testing.T) {
	if any((*Application).ReceiveLegacyArchiveHit) == nil {
		t.Fatal("legacy archive compatibility hit operation is unavailable")
	}
}
