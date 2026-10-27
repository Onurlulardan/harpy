package portaluserpostgres

import _ "embed"

//go:embed sql/tenant/create.sql
var createSQL string

//go:embed sql/tenant/change_status.sql
var changeStatusSQL string

//go:embed sql/tenant/grant_console.sql
var grantConsoleSQL string

//go:embed sql/tenant/revoke_console.sql
var revokeConsoleSQL string

//go:embed sql/tenant/read.sql
var readSQL string

//go:embed sql/tenant/project_subject_links.sql
var projectSubjectLinksSQL string

//go:embed sql/tenant/verify_access.sql
var verifyAccessSQL string

//go:embed sql/platform/directory_upsert.sql
var directoryUpsertSQL string
