package licenseauthority

import _ "embed"

var (
	//go:embed sql/platform/claim_refresh.sql
	claimRefreshSQL string
	//go:embed sql/platform/accept_snapshot.sql
	acceptSnapshotSQL string
	//go:embed sql/platform/record_refresh_failure.sql
	recordRefreshFailureSQL string
	//go:embed sql/platform/commit_restored_license.sql
	commitRestoredLicenseSQL string
	//go:embed sql/platform/read_runtime.sql
	readRuntimeSQL string
	//go:embed sql/platform/acquire_refresh_target.sql
	acquireRefreshTargetSQL string
	//go:embed sql/platform/read_restore_binding.sql
	readRestoreBindingSQL string
	//go:embed sql/platform/read_restored_receipt.sql
	readRestoredLicenseReceiptSQL string
)
