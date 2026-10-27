package subscription

import _ "embed"

var (
	//go:embed sql/platform/resolve_snapshot_authority.sql
	resolveSnapshotAuthoritySQL string
	//go:embed sql/platform/get_current.sql
	getCurrentSQL string
)
