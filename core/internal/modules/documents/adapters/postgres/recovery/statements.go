package recovery

import _ "embed"

var (
	//go:embed sql/tenant/begin_hold.sql
	beginHoldSQL string
	//go:embed sql/tenant/inventory.sql
	inventorySQL string
	//go:embed sql/tenant/record_copy.sql
	recordCopySQL string
	//go:embed sql/tenant/seal_hold.sql
	sealHoldSQL string
	//go:embed sql/tenant/release_hold.sql
	releaseHoldSQL string
	//go:embed sql/tenant/fail_hold.sql
	failHoldSQL string
	//go:embed sql/tenant/apply_locator.sql
	applyLocatorSQL string
	//go:embed sql/tenant/terminalize_transients.sql
	terminalizeTransientsSQL string
	//go:embed sql/tenant/import_snapshot.sql
	importSnapshotSQL string
)
