package systemtransfer

import _ "embed"

//go:embed sql/tenant/register_source.sql
var registerSourceSQL string

//go:embed sql/tenant/submit_export.sql
var submitExportSQL string

//go:embed sql/tenant/read_export.sql
var readExportSQL string

//go:embed sql/tenant/read_export_artifact.sql
var readExportArtifactSQL string

//go:embed sql/tenant/cancel_export.sql
var cancelExportSQL string

//go:embed sql/tenant/submit_import.sql
var submitImportSQL string

//go:embed sql/tenant/read_import.sql
var readImportSQL string

//go:embed sql/tenant/cancel_import.sql
var cancelImportSQL string

//go:embed sql/tenant/read_receipts.sql
var readReceiptsSQL string
