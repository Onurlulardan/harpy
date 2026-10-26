package offering

import _ "embed"

var (
	//go:embed sql/platform/resolve_authority.sql
	resolveAuthoritySQL string
	//go:embed sql/platform/create.sql
	createSQL string
	//go:embed sql/platform/publish.sql
	publishSQL string
	//go:embed sql/platform/retire.sql
	retireSQL string
	//go:embed sql/platform/add_price.sql
	addPriceSQL string
	//go:embed sql/platform/add_grant.sql
	addGrantSQL string
)
