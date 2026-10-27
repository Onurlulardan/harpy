package knowledgefile

import _ "embed"

//go:embed sql/tenant/create.sql
var createSQL string

//go:embed sql/tenant/complete_stage.sql
var completeStageSQL string

//go:embed sql/tenant/prepare_adopt.sql
var prepareAdoptSQL string

//go:embed sql/tenant/complete_adopt.sql
var completeAdoptSQL string

//go:embed sql/tenant/prepare_release.sql
var prepareReleaseSQL string

//go:embed sql/tenant/complete_release.sql
var completeReleaseSQL string

//go:embed sql/tenant/verify_open.sql
var verifyOpenSQL string

//go:embed sql/tenant/expire.sql
var expireSQL string
