package postgres

import _ "embed"

var (
	//go:embed sql/tenant/create.sql
	createSQL string
	//go:embed sql/tenant/get.sql
	getSQL string
	//go:embed sql/tenant/list.sql
	listSQL string
	//go:embed sql/tenant/update.sql
	updateSQL string
	//go:embed sql/tenant/delete.sql
	deleteSQL string
	//go:embed sql/tenant/publish.sql
	publishSQL string
	//go:embed sql/tenant/resolve.sql
	resolveSQL string
)
