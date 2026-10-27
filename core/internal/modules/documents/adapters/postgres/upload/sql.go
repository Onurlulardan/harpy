package upload

import _ "embed"

var (
	//go:embed sql/tenant/find_create.sql
	findCreateSQL string
	//go:embed sql/tenant/create.sql
	createSQL string
	//go:embed sql/tenant/get.sql
	getSQL string
	//go:embed sql/tenant/begin.sql
	beginSQL string
	//go:embed sql/tenant/record_effect.sql
	recordEffectSQL string
	//go:embed sql/tenant/record_failure.sql
	recordFailureSQL string
	//go:embed sql/tenant/prepare_finalize.sql
	prepareFinalizeSQL string
	//go:embed sql/tenant/finalize.sql
	finalizeSQL string
	//go:embed sql/tenant/cancel.sql
	cancelSQL string
)
