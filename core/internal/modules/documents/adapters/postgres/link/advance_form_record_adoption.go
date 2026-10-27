package link

import (
	"context"
	"time"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
)

func (storage *store) AdvanceFormRecordAdoption(
	ctx context.Context,
	expected linkapp.FormRecordAdoptionCheckpoint,
	next linkapp.FormRecordAdoptionCheckpoint,
	updatedAt time.Time,
) (linkapp.FormRecordAdoptionCheckpoint, error) {
	if !storage.usable(ctx) || !validAdoptionCheckpoint(expected, true) ||
		!validAdoptionCheckpoint(next, true) || expected.OperationID != next.OperationID ||
		expected.DefinitionID != next.DefinitionID || expected.Epoch != next.Epoch ||
		expected.Done || next.RecordCount < expected.RecordCount ||
		next.LinkCount < expected.LinkCount {
		return linkapp.FormRecordAdoptionCheckpoint{}, linkapp.ErrInvalid
	}
	return scanFormRecordAdoption(storage.queryRow(ctx, advanceFormRecordAdoptionSQL,
		expected.OperationID.String(), expected.DefinitionID.String(), expected.Epoch,
		adoptionCursor(expected.Cursor), expected.RecordCount, expected.LinkCount,
		expected.SourceSHA256[:], expected.LinkSHA256[:],
		adoptionCursor(next.Cursor), next.RecordCount, next.LinkCount,
		next.SourceSHA256[:], next.LinkSHA256[:], next.Done, updatedAt))
}
