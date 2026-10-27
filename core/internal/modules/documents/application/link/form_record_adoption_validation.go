package link

import (
	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func validAdoptionPage(page formrecordlink.AdoptionPage) bool {
	if !domain.ID(page.OperationID).UUIDv7() ||
		!domain.ID(page.DefinitionID).UUIDv7() || page.Epoch < 2 ||
		len(page.Entries) > formrecordlink.AdoptionPageLimit ||
		page.AppliedAt.IsZero() || page.AppliedAt.Location().String() != "UTC" ||
		page.AppliedAt.Nanosecond()%1_000 != 0 ||
		page.Previous.OperationID != page.OperationID ||
		page.Previous.DefinitionID != page.DefinitionID ||
		page.Previous.Epoch != page.Epoch || page.Previous.Done ||
		!page.Previous.Valid() {
		return false
	}
	progress := page.Previous
	for _, entry := range page.Entries {
		next, ok := formrecordlink.AdvanceAdoption(progress, entry)
		if !ok {
			return false
		}
		progress = next
	}
	return len(page.Entries) > 0 || page.Done
}

func adoptionSeed(
	operationID, definitionID formrecordlink.ID, epoch int64,
) FormRecordAdoptionCheckpoint {
	source, links := formrecordlink.AdoptionSeeds(definitionID, epoch)
	return FormRecordAdoptionCheckpoint{OperationID: domain.ID(operationID),
		DefinitionID: domain.ID(definitionID), Epoch: epoch,
		SourceSHA256: source, LinkSHA256: links}
}

func adoptionCheckpoint(
	value formrecordlink.AdoptionProgress,
) FormRecordAdoptionCheckpoint {
	return FormRecordAdoptionCheckpoint{OperationID: domain.ID(value.OperationID),
		DefinitionID: domain.ID(value.DefinitionID), Cursor: domain.ID(value.Cursor),
		Epoch: value.Epoch, RecordCount: value.RecordCount,
		LinkCount: value.LinkCount, SourceSHA256: value.SourceSHA256,
		LinkSHA256: value.LinkSHA256, Done: value.Done}
}

func exportAdoption(
	value FormRecordAdoptionCheckpoint,
) formrecordlink.AdoptionProgress {
	return formrecordlink.AdoptionProgress{OperationID: formrecordlink.ID(value.OperationID),
		DefinitionID: formrecordlink.ID(value.DefinitionID),
		Cursor:       formrecordlink.ID(value.Cursor), Epoch: value.Epoch,
		RecordCount: value.RecordCount, LinkCount: value.LinkCount,
		SourceSHA256: value.SourceSHA256, LinkSHA256: value.LinkSHA256,
		Done: value.Done}
}
