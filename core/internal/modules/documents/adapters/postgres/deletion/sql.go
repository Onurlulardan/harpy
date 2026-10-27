package deletion

import _ "embed"

var (
	//go:embed sql/tenant/claim.sql
	claimSQL string
	//go:embed sql/tenant/begin_reconcile.sql
	beginReconcileSQL string
	//go:embed sql/tenant/confirm.sql
	confirmSQL string
)
