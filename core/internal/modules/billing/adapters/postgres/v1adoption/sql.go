package v1adoption

import _ "embed"

//go:embed sql/platform/register_scope.sql
var registerScopeSQL string

//go:embed sql/platform/apply_scope.sql
var applyScopeSQL string

//go:embed sql/platform/verify_scope.sql
var verifyScopeSQL string

//go:embed sql/platform/rehearse_rollback.sql
var rehearseRollbackSQL string
