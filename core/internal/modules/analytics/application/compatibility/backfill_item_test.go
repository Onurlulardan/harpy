package compatibility

import (
	"context"
	"testing"
)

func TestBackfillItemPersistsExactAdoptedIdentity(t *testing.T) {
	application, store := applicationFixture(t, mappingFixture())
	item := AdoptionItem{RunID: runID, LegacyEndpointID: legacyID,
		SourceSHA256: [32]byte{1}, TargetSHA256: [32]byte{2},
		Disposition: DispositionAdopted, QueryID: queryID, QueryVersionID: versionID,
		PublicationID: releaseID}
	result, err := application.BackfillItem(context.Background(), authority, item)
	if err != nil || result.CreatedAt != compatNow || store.item != result {
		t.Fatalf("result=%+v stored=%+v error=%v", result, store.item, err)
	}
}

func TestBackfillItemRejectsSilentBlockedDowngrade(t *testing.T) {
	application, _ := applicationFixture(t, mappingFixture())
	item := AdoptionItem{RunID: runID, LegacyEndpointID: legacyID,
		SourceSHA256: [32]byte{1}, Disposition: DispositionBlocked}
	if _, err := application.BackfillItem(context.Background(), authority, item); err != ErrInvalid {
		t.Fatalf("error=%v", err)
	}
}
