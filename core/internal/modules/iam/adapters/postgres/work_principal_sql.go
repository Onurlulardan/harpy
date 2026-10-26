package postgres

import _ "embed"

//go:embed sql/tenant/work_principal_evidence.sql
var workPrincipalEvidenceSQL string

//go:embed sql/tenant/work_available_users.sql
var workAvailableUsersSQL string
