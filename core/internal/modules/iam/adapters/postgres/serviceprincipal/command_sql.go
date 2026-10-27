package serviceprincipal

import _ "embed"

//go:embed sql/tenant/create.sql
var createSQL string

//go:embed sql/tenant/disable.sql
var disableSQL string

//go:embed sql/tenant/enable.sql
var enableSQL string

//go:embed sql/tenant/retire.sql
var retireSQL string

//go:embed sql/tenant/grant.sql
var grantSQL string

//go:embed sql/tenant/revoke.sql
var revokeSQL string
