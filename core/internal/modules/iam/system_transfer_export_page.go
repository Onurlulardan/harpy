package iam

import (
	"encoding/json"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func systemTransferExportRecords(
	items []json.RawMessage,
	fields []string,
) ([]systemTransferExportRecord, error) {
	records := make([]systemTransferExportRecord, len(items))
	pageIDs := make(map[uuid.UUID]struct{}, len(items))
	for index, item := range items {
		record, err := decodeSystemTransferExportRecord(item, fields)
		if err != nil {
			return nil, err
		}
		if _, duplicate := pageIDs[record.id]; duplicate {
			return nil, systemtransfer.ErrExportStateConflict
		}
		pageIDs[record.id] = struct{}{}
		records[index] = record
	}
	return records, nil
}

func (source *systemTransferExportRows) firstPage(
	records []systemTransferExportRecord,
	hasMore bool,
) ([]uuid.UUID, [][]systemtransfer.ExportCell, bool, error) {
	if hasMore && len(records) == 0 {
		return nil, nil, false, systemtransfer.ErrExportStateConflict
	}
	identifiers := make([]uuid.UUID, len(records))
	rows := make([][]systemtransfer.ExportCell, len(records))
	for index, record := range records {
		if _, duplicate := source.seen[record.id]; duplicate {
			return nil, nil, false, systemtransfer.ErrExportStateConflict
		}
		source.seen[record.id] = struct{}{}
		identifiers[index] = record.id
		rows[index] = make([]systemtransfer.ExportCell, len(source.plan.columns))
	}
	return identifiers, rows, hasMore, nil
}

func systemTransferExportSamePage(
	identifiers []uuid.UUID,
	records []systemTransferExportRecord,
	expectedMore bool,
	actualMore bool,
) bool {
	if expectedMore != actualMore || len(identifiers) != len(records) {
		return false
	}
	for index := range records {
		if identifiers[index] != records[index].id {
			return false
		}
	}
	return true
}

func (source *systemTransferExportRows) applyProjection(
	rows [][]systemtransfer.ExportCell,
	records []systemTransferExportRecord,
	projection systemTransferExportProjection,
) error {
	for rowIndex, record := range records {
		for _, field := range projection.fields {
			position, selected := source.positions[field]
			if !selected {
				if field != "id" {
					return systemtransfer.ErrExportStateConflict
				}
				continue
			}
			cell, err := systemTransferExportCell(
				source.plan.kinds[position], record.members[field],
			)
			if err != nil {
				return err
			}
			if rows[rowIndex][position].Kind != "" && rows[rowIndex][position] != cell {
				return systemtransfer.ErrExportStateConflict
			}
			rows[rowIndex][position] = cell
		}
	}
	return nil
}

func advanceSystemTransferProjection(
	projection *systemTransferExportProjection,
	result queryapp.Result,
) error {
	if projection == nil {
		return systemtransfer.ErrExportStateConflict
	}
	if result.HasMore() {
		if result.NextCursor() == "" || result.NextCursor() == projection.cursor {
			return systemtransfer.ErrExportStateConflict
		}
		projection.cursor, projection.more = result.NextCursor(), true
		return nil
	}
	if result.NextCursor() != "" {
		return systemtransfer.ErrExportStateConflict
	}
	projection.cursor, projection.more = "", false
	return nil
}
