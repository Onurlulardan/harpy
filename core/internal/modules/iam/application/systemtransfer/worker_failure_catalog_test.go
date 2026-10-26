package systemtransfer

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestImportCatalogPinDriftIsAStablePermanentFailure(t *testing.T) {
	actorID := uuid.MustParse("019be400-0000-7000-8000-000000000001")
	err := ClassifyWorkerFailure(ErrImportCatalogPinStale, actorID)
	var failure *WorkerFailure
	if !errors.As(err, &failure) || failure.Class != WorkerFailurePermanent ||
		failure.Code != "IMPORT_CATALOG_PIN_STALE" || failure.ActorID != actorID ||
		failure.Digest == [32]byte{} {
		t.Fatalf("failure=%+v err=%v", failure, err)
	}
}
