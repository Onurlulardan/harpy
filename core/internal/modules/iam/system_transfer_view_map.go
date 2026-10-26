package iam

import (
	"encoding/hex"
	"fmt"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func systemExportView(record systemtransfer.ExportRecord) SystemExportView {
	resource, _ := ParseSystemTransferResource(record.Resource)
	return SystemExportView{
		ExportID: record.ExportID, Resource: resource, Format: record.Format,
		State: record.State, Rows: record.Rows, Bytes: record.Bytes,
		CreatedAt: record.CreatedAt, ReadyAt: record.ReadyAt,
		ExpiresAt: record.ExpiresAt, Replay: record.Replay,
	}
}

func systemImportSourceView(record systemtransfer.SourceRecord) SystemImportSourceView {
	resource, _ := ParseSystemTransferResource(record.Resource)
	columns := make([]SystemImportSourceColumnView, len(record.Analysis.Columns))
	for index, column := range record.Analysis.Columns {
		columns[index] = SystemImportSourceColumnView{Name: column.Name, Kind: column.Kind}
	}
	return SystemImportSourceView{
		SourceID: record.SourceID, Resource: resource, Format: record.Format,
		Bytes: record.PlaintextBytes, SHA256: record.PlaintextSHA256,
		UploadedAt: record.UploadedAt, Replay: record.Replay,
		Columns: columns, Samples: record.Analysis.OwnedCopy().Samples,
	}
}

func systemImportView(record systemtransfer.ImportRecord) SystemImportView {
	resource, _ := ParseSystemTransferResource(record.Resource)
	return SystemImportView{
		ImportID: record.ImportID, Resource: resource, State: record.State,
		RowsSeen: record.RowsSeen, Created: record.Created, Updated: record.Updated,
		Skipped: record.Skipped, Failed: record.Failed, CreatedAt: record.CreatedAt,
		CompletedAt: record.CompletedAt, Replay: record.Replay,
	}
}

func systemImportReportView(
	importID uuid.UUID,
	page systemtransfer.ReceiptPage,
) SystemImportReportView {
	receipts := make([]SystemImportReceiptView, 0, len(page.Rows))
	for _, row := range page.Rows {
		ownerID := ""
		if row.OwnerID != nil {
			ownerID = row.OwnerID.String()
		}
		receipts = append(receipts, SystemImportReceiptView{
			RowIndex: row.RowIndex, Outcome: row.Outcome, OwnerKind: row.OwnerKind,
			OwnerID: ownerID, ErrorCode: row.ErrorCode,
			SourceDigest: hex.EncodeToString(row.SourceSHA256[:]),
			ResultDigest: hex.EncodeToString(row.ResultSHA256[:]),
		})
	}
	return SystemImportReportView{
		ImportID: importID, Receipts: receipts,
		NextAfterRow: page.NextAfter, HasMore: page.HasMore,
	}
}

func systemExportContentMetadata(resource SystemTransferResource,
	record systemtransfer.ExportArtifact) (string, string, error) {
	var contentType string
	switch record.Format {
	case "csv":
		contentType = "text/csv; charset=utf-8"
	case "jsonl":
		contentType = "application/x-ndjson"
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "", "", ErrSystemTransferUnavailable
	}
	name := fmt.Sprintf("iam-%s-%s.%s", resource, record.ExportID, record.Format)
	return name, contentType, nil
}
