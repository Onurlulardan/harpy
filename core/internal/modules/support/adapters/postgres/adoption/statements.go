package adoption

import _ "embed"

var (
	//go:embed sql/platform/start.sql
	startSQL string
	//go:embed sql/platform/claim.sql
	claimSQL string
	//go:embed sql/platform/apply_contact.sql
	applyContactSQL string
	//go:embed sql/platform/apply_bug.sql
	applyBugSQL string
	//go:embed sql/platform/advance.sql
	advanceSQL string
	//go:embed sql/platform/reconcile.sql
	reconcileSQL string
	//go:embed sql/platform/complete.sql
	completeSQL string
)
