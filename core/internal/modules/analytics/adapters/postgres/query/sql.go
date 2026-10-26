package query

import _ "embed"

var (
	//go:embed sql/tenant/create.sql
	createSQL string
	//go:embed sql/tenant/get.sql
	getSQL string
	//go:embed sql/tenant/latest_published.sql
	latestPublishedSQL string
	//go:embed sql/tenant/get_version.sql
	getVersionSQL string
	//go:embed sql/tenant/resolve_query_id.sql
	resolveQueryIDSQL string
	//go:embed sql/tenant/dependencies.sql
	dependenciesSQL string
	//go:embed sql/tenant/list.sql
	listSQL string
	//go:embed sql/tenant/update.sql
	updateSQL string
	//go:embed sql/tenant/delete.sql
	deleteSQL string
	//go:embed sql/tenant/publish.sql
	publishSQL string
	//go:embed sql/tenant/revalidate.sql
	revalidateSQL string
)
