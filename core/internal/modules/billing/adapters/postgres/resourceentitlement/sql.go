package resourceentitlement

import _ "embed"

var (
	//go:embed sql/platform/resolve_authority.sql
	resolveAuthoritySQL string
	//go:embed sql/platform/assert_authority.sql
	assertAuthoritySQL string
	//go:embed sql/platform/inspect.sql
	inspectSQL string
	//go:embed sql/platform/resolve_feature.sql
	resolveFeatureSQL string
	//go:embed sql/platform/record.sql
	recordSQL string
	//go:embed sql/platform/inspect_reconciliation.sql
	inspectReconciliationSQL string
	//go:embed sql/platform/record_reconciliation.sql
	recordReconciliationSQL string
)
