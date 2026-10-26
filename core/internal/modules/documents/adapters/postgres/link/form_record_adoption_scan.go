package link

import (
	"time"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func scanFormRecordAdoption(value row) (linkapp.FormRecordAdoptionCheckpoint, error) {
	var operation, definition, cursor string
	var epoch, records, links int64
	var source, catalog []byte
	var done bool
	if value == nil || value.Scan(&operation, &definition, &epoch, &cursor,
		&records, &links, &source, &catalog, &done) != nil {
		return linkapp.FormRecordAdoptionCheckpoint{}, linkapp.ErrUnavailable
	}
	operationID, operationErr := domain.ParseID(operation)
	definitionID, definitionErr := domain.ParseID(definition)
	var cursorID domain.ID
	var cursorErr error
	if cursor != "00000000-0000-0000-0000-000000000000" {
		cursorID, cursorErr = domain.ParseID(cursor)
	}
	if operationErr != nil || definitionErr != nil || cursorErr != nil || len(source) != 32 ||
		len(catalog) != 32 {
		return linkapp.FormRecordAdoptionCheckpoint{}, linkapp.ErrUnavailable
	}
	result := linkapp.FormRecordAdoptionCheckpoint{OperationID: operationID,
		DefinitionID: definitionID, Cursor: cursorID, Epoch: epoch,
		RecordCount: records, LinkCount: links, Done: done}
	copy(result.SourceSHA256[:], source)
	copy(result.LinkSHA256[:], catalog)
	return result, nil
}

func scanFormRecordAdoptionReceipt(
	value row,
) (linkapp.FormRecordAdoptionReceipt, error) {
	var source, receipt []byte
	var issued time.Time
	if value == nil || value.Scan(&source, &receipt, &issued) != nil ||
		len(source) != 32 || len(receipt) != 32 {
		return linkapp.FormRecordAdoptionReceipt{}, linkapp.ErrUnavailable
	}
	result := linkapp.FormRecordAdoptionReceipt{IssuedAt: issued.UTC()}
	copy(result.SourceReceiptSHA256[:], source)
	copy(result.ReceiptSHA256[:], receipt)
	return result, nil
}
