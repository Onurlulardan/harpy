package postgres

import _ "embed"

//go:embed sql/platform/resolve_login_credential_target.sql
var resolveLoginCredentialTargetSQL string

//go:embed sql/platform/resolve_refresh_credential_target.sql
var resolveRefreshCredentialTargetSQL string

//go:embed sql/platform/resolve_tenant_credential_target.sql
var resolveTenantCredentialTargetSQL string
