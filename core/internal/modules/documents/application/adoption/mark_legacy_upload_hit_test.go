package adoption

import "testing"

func TestMarkLegacyUploadHitOperationExists(t *testing.T) {
	if any((*Application).MarkLegacyUploadHit) == nil {
		t.Fatal("legacy upload compatibility hit operation is unavailable")
	}
}
