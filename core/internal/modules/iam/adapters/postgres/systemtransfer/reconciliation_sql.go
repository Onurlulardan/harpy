package systemtransfer

import _ "embed"

var (
	//go:embed sql/tenant/claim_export_reconciliation.sql
	reconciliationAcquireSQL string
	//go:embed sql/tenant/resolve_export_reconciliation_sealed.sql
	reconciliationSealedSQL string
	//go:embed sql/tenant/resolve_export_reconciliation_missing.sql
	reconciliationMissingSQL string
	//go:embed sql/tenant/retry_export_reconciliation.sql
	reconciliationRetrySQL string
	//go:embed sql/tenant/park_export_reconciliation.sql
	reconciliationParkedSQL string
)
