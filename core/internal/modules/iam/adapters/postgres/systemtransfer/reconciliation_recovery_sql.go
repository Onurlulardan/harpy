package systemtransfer

import _ "embed"

//go:embed sql/tenant/recover_expired_export_reconciliation.sql
var recoverExpiredExportReconciliationSQL string
