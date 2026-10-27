package jobs

import _ "embed"

var (
	//go:embed sql/tenant/cache_warm.sql
	cacheWarmSQL string
	//go:embed sql/tenant/query_revalidate.sql
	queryRevalidateSQL string
	//go:embed sql/tenant/publication_repair.sql
	publicationRepairSQL string
)
