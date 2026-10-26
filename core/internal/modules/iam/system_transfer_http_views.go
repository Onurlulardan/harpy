package iam

import "encoding/hex"

func systemExportHTTPView(view SystemExportView) map[string]any {
	return map[string]any{"export_id": view.ExportID, "resource": view.Resource,
		"format": view.Format, "state": view.State, "rows": view.Rows,
		"bytes": view.Bytes, "created_at": view.CreatedAt, "ready_at": view.ReadyAt,
		"expires_at": view.ExpiresAt, "replay": view.Replay}
}

func systemImportSourceHTTPView(view SystemImportSourceView) map[string]any {
	columns := make([]map[string]any, len(view.Columns))
	for index, column := range view.Columns {
		columns[index] = map[string]any{"name": column.Name, "kind": column.Kind}
	}
	return map[string]any{"source_id": view.SourceID, "resource": view.Resource,
		"format": view.Format, "bytes": view.Bytes,
		"sha256": hex.EncodeToString(view.SHA256[:]), "uploaded_at": view.UploadedAt,
		"replay": view.Replay, "columns": columns, "samples": view.Samples}
}

func systemImportHTTPView(view SystemImportView) map[string]any {
	return map[string]any{"import_id": view.ImportID, "resource": view.Resource,
		"state": view.State, "rows_seen": view.RowsSeen, "created": view.Created,
		"updated": view.Updated, "skipped": view.Skipped, "failed": view.Failed,
		"created_at": view.CreatedAt, "completed_at": view.CompletedAt,
		"replay": view.Replay}
}
