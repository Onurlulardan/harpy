package serviceprincipal

import _ "embed"

//go:embed sql/tenant/evidence.sql
var evidenceSQL string
