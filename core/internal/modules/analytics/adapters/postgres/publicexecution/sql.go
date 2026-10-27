package publicexecution

import _ "embed"

var (
	//go:embed sql/tenant/query_root.sql
	queryRootSQL string
	//go:embed sql/tenant/dashboard_root.sql
	dashboardRootSQL string
)
