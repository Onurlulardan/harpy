package aiplatformowner

import _ "embed"

//go:embed sql/tenant/current_session.sql
var currentSessionSQL string

//go:embed sql/tenant/current_owner.sql
var currentOwnerSQL string
