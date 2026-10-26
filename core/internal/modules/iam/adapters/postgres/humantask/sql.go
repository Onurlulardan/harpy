package humantask

import _ "embed"

//go:embed sql/tenant/resolve_candidate.sql
var resolveCandidateSQL string

//go:embed sql/tenant/resolve_assignment.sql
var resolveAssignmentSQL string
