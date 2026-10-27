package initialadminsetup

import _ "embed"

//go:embed sql/platform/stage.sql
var stageSQL string

//go:embed sql/tenant/finalize.sql
var finalizeSQL string

//go:embed sql/tenant/reissue.sql
var reissueSQL string
