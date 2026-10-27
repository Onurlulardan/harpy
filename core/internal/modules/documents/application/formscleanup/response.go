package formscleanup

import (
	"time"

	formscleanupapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func snapshotMatches(snapshot Snapshot, operationID, definitionID domain.ID,
	epoch int64, requestedAt time.Time,
) bool {
	return snapshot.OperationID == operationID && snapshot.DefinitionID == definitionID &&
		snapshot.Epoch == epoch && snapshot.RequestedAt == requestedAt
}

func exportProgress(snapshot Snapshot) (formscleanupapi.Progress, error) {
	if !snapshot.OperationID.UUIDv7() || !snapshot.DefinitionID.UUIDv7() ||
		snapshot.Epoch < 2 {
		return formscleanupapi.Progress{}, ErrUnavailable
	}
	progress := formscleanupapi.Progress{OperationID: snapshot.OperationID.String(),
		DefinitionID: snapshot.DefinitionID.String(), Epoch: snapshot.Epoch,
		State: snapshot.State}
	if snapshot.State != formscleanupapi.StateSucceeded {
		return progress, nil
	}
	if snapshot.LinkCatalogSHA256 == [32]byte{} ||
		snapshot.ZeroBlockerSHA256 == [32]byte{} ||
		snapshot.ReceiptSHA256 == [32]byte{} || snapshot.CompletedAt.IsZero() {
		return formscleanupapi.Progress{}, ErrUnavailable
	}
	progress.Receipt = &formscleanupapi.Receipt{
		OperationID: snapshot.OperationID.String(), DefinitionID: snapshot.DefinitionID.String(),
		Epoch: snapshot.Epoch, CatalogVersion: formscleanupapi.CatalogVersion,
		LinkCount: snapshot.LinkCount, ItemCount: snapshot.ItemCount,
		VersionCount: snapshot.VersionCount, CiphertextBytes: snapshot.CiphertextBytes,
		LinkCatalogSHA256: snapshot.LinkCatalogSHA256,
		ZeroBlockerSHA256: snapshot.ZeroBlockerSHA256,
		ReceiptSHA256:     snapshot.ReceiptSHA256, IssuedAt: snapshot.CompletedAt,
	}
	return progress, nil
}
