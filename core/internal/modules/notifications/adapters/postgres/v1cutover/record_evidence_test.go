package v1cutover

import (
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func TestRecordEvidenceRejectsInvalidInventory(t *testing.T) {
	if _, err := (&Store{}).RecordEvidence(t.Context(), app.Inventory{}); err != app.ErrInvalid {
		t.Fatalf("RecordEvidence() error = %v", err)
	}
}
