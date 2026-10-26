package recoveryanchorpostgres

import _ "embed"

//go:embed sql/platform/finalize_authority.sql
var finalizeAuthoritySQL string
