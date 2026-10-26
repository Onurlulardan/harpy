package restoreexecution

import _ "embed"

var (
	//go:embed sql/platform/claim.sql
	claimSQL string
	//go:embed sql/platform/heartbeat.sql
	heartbeatSQL string
	//go:embed sql/platform/begin_drain.sql
	beginDrainSQL string
	//go:embed sql/platform/begin_apply.sql
	beginApplySQL string
	//go:embed sql/platform/begin_verify.sql
	beginVerifySQL string
	//go:embed sql/platform/mark_switched.sql
	markSwitchedSQL string
	//go:embed sql/platform/complete.sql
	completeSQL string
	//go:embed sql/platform/fail.sql
	failSQL string
	//go:embed sql/platform/import_receipt.sql
	importReceiptSQL string

	//go:embed sql/platform/seed_rehydrate.sql
	seedRehydrateSQL string
)
