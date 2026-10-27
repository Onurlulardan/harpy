package primarytarget

import _ "embed"

//go:embed sql/platform/get.sql
var getPrimaryTenantTargetSQL string
