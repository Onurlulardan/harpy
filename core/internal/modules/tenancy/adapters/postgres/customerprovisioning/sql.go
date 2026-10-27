package customerprovisioning

import _ "embed"

//go:embed sql/platform/accept.sql
var acceptSQL string

//go:embed sql/platform/get.sql
var getSQL string

//go:embed sql/platform/get_setup_reissue_target.sql
var getSetupReissueTargetSQL string
