package systemtransfercleanup

import _ "embed"

//go:embed sql/platform/next_tenant.sql
var nextTenantSQL string

//go:embed sql/tenant/claim.sql
var claimSQL string

//go:embed sql/tenant/complete.sql
var completeSQL string

//go:embed sql/tenant/retry.sql
var retrySQL string
