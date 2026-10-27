package compatibility

import (
	"context"
	"testing"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	"github.com/google/uuid"
)

func TestBackfillItemUsesAdopterAuthorityAndExactReplayResult(t *testing.T) {
	item := compatibilityapp.AdoptionItem{RunID: testRun, LegacyEndpointID: testLegacy,
		SourceSHA256: [32]byte{1}, TargetSHA256: [32]byte{2},
		Disposition: compatibilityapp.DispositionAdopted, QueryID: testQuery,
		QueryVersionID: testVersion, PublicationID: testRelease, CreatedAt: testTime}
	store, database := newStore(t, scanRow(func(targets []any) error {
		*targets[0].(*uuid.UUID) = item.RunID
		*targets[1].(*uuid.UUID) = item.LegacyEndpointID
		*targets[2].(*[]byte) = item.SourceSHA256[:]
		*targets[3].(*string) = string(item.Disposition)
		*targets[4].(**string) = nil
		*targets[5].(**uuid.UUID) = &item.QueryID
		*targets[6].(**uuid.UUID) = &item.QueryVersionID
		*targets[7].(**uuid.UUID) = &item.PublicationID
		*targets[8].(*[]byte) = item.TargetSHA256[:]
		*targets[9].(*time.Time) = item.CreatedAt
		*targets[10].(*bool) = true
		return nil
	}))
	result, replay, err := store.BackfillItem(context.Background(), testAuthority, item)
	if err != nil || !replay || result.QueryID != testQuery || len(database.args) != 13 ||
		database.args[0] != testAuthority.SchemaName {
		t.Fatalf("result=%+v replay=%v args=%v error=%v", result, replay, database.args, err)
	}
}
