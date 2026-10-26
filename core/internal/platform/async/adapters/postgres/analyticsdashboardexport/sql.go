package analyticsdashboardexport

import _ "embed"

//go:embed sql/candidate.sql
var candidateSQL string

//go:embed sql/unknown_candidate.sql
var unknownCandidateSQL string

//go:embed sql/maintenance_unknown_candidate.sql
var maintenanceUnknownCandidateSQL string

//go:embed sql/expired_candidate.sql
var expiredCandidateSQL string

//go:embed sql/spool_tenant.sql
var spoolTenantSQL string
