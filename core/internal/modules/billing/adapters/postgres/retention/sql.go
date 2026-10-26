package retention

import _ "embed"

//go:embed sql/platform/resolve_authority.sql
var resolveAuthoritySQL string

//go:embed sql/platform/execute_cycle.sql
var executeCycleSQL string
