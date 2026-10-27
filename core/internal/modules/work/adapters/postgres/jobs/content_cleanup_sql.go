package jobs

import _ "embed"

//go:embed sql/tenant/execute_purge.sql
var executePurgeSQL string
