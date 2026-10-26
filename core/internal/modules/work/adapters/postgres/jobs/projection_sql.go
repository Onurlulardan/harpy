package jobs

import _ "embed"

//go:embed sql/tenant/repair_projection.sql
var repairProjectionSQL string
