package link

import (
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func validAdoptionCheckpoint(
	value linkapp.FormRecordAdoptionCheckpoint,
	cursorAllowed bool,
) bool {
	if !value.OperationID.UUIDv7() || !value.DefinitionID.UUIDv7() ||
		value.Epoch < 2 || value.RecordCount < 0 || value.LinkCount < 0 ||
		value.SourceSHA256 == [32]byte{} || value.LinkSHA256 == [32]byte{} {
		return false
	}
	if value.RecordCount == 0 {
		return value.Cursor == (domain.ID{})
	}
	return cursorAllowed && value.Cursor.UUIDv7()
}

func adoptionCursor(value domain.ID) string {
	if value == (domain.ID{}) {
		return "00000000-0000-0000-0000-000000000000"
	}
	return value.String()
}
