package retention

import _ "embed"

//go:embed sql/tenant/execute.sql
var executeCycleSQL string
