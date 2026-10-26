package adoption

import _ "embed"

var (
	//go:embed sql/tenant/create_run.sql
	createRunSQL string
	//go:embed sql/tenant/inventory.sql
	inventorySQL string
	//go:embed sql/tenant/prepare_copy.sql
	prepareCopySQL string
	//go:embed sql/tenant/verify_copy.sql
	verifyCopySQL string
	//go:embed sql/tenant/begin_shadow.sql
	beginShadowSQL string
	//go:embed sql/tenant/reconcile.sql
	reconcileSQL string
	//go:embed sql/tenant/accept.sql
	acceptSQL string
	//go:embed sql/tenant/record_hit.sql
	recordHitSQL string
	//go:embed sql/tenant/request_reverse.sql
	requestReverseSQL string
	//go:embed sql/tenant/decide_reverse.sql
	decideReverseSQL string
)
