package discounteligibility

import _ "embed"

var (
	//go:embed sql/platform/resolve_authority.sql
	resolveAuthoritySQL string
	//go:embed sql/platform/find_eligible.sql
	findEligibleSQL string
	//go:embed sql/platform/list_offerings.sql
	listOfferingsSQL string
)
