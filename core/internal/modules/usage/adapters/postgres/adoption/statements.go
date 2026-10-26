package adoption

import _ "embed"

var (
	//go:embed sql/platform/start.sql
	startSQL string
	//go:embed sql/platform/claim.sql
	claimSQL string
	//go:embed sql/platform/apply.sql
	applySQL string
	//go:embed sql/platform/advance.sql
	advanceSQL string
	//go:embed sql/platform/reconcile.sql
	reconcileSQL string
	//go:embed sql/platform/complete.sql
	completeSQL string
)
