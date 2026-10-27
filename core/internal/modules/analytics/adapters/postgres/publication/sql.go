package publication

import _ "embed"

var (
	//go:embed sql/tenant/create.sql
	createSQL string
	//go:embed sql/tenant/get.sql
	getSQL string
	//go:embed sql/tenant/list.sql
	listSQL string
	//go:embed sql/tenant/submit.sql
	submitSQL string
	//go:embed sql/tenant/approve.sql
	approveSQL string
	//go:embed sql/tenant/activate.sql
	activateSQL string
	//go:embed sql/tenant/revoke.sql
	revokeSQL string
)
