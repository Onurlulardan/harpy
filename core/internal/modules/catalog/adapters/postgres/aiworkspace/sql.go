package aiworkspace

import _ "embed"

var (
	//go:embed sql/tenant/list.sql
	listSQL string
	//go:embed sql/tenant/search.sql
	searchSQL string
	//go:embed sql/tenant/describe.sql
	describeSQL string
	//go:embed sql/tenant/recent.sql
	recentSQL string
	//go:embed sql/tenant/history.sql
	historySQL string
)
