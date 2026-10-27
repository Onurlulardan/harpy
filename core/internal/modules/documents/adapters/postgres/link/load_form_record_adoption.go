package link

import (
	"context"
	"time"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
)

func (storage *store) LoadFormRecordAdoption(
	ctx context.Context,
	seed linkapp.FormRecordAdoptionCheckpoint,
	startedAt time.Time,
) (linkapp.FormRecordAdoptionCheckpoint, error) {
	if !storage.usable(ctx) || !validAdoptionCheckpoint(seed, false) ||
		seed.RecordCount != 0 || seed.LinkCount != 0 || seed.Done {
		return linkapp.FormRecordAdoptionCheckpoint{}, linkapp.ErrInvalid
	}
	return scanFormRecordAdoption(storage.queryRow(ctx, loadFormRecordAdoptionSQL,
		seed.OperationID.String(), seed.DefinitionID.String(), seed.Epoch,
		seed.SourceSHA256[:], seed.LinkSHA256[:], startedAt))
}
