package discount

import _ "embed"

var (
	//go:embed sql/platform/resolve_authority.sql
	resolveAuthoritySQL string
	//go:embed sql/platform/assert_authority.sql
	assertAuthoritySQL string
	//go:embed sql/platform/lock_eligibility.sql
	lockEligibilitySQL string
	//go:embed sql/platform/lock_redemption.sql
	lockRedemptionSQL string
	//go:embed sql/platform/find_redemption.sql
	findRedemptionSQL string
	//go:embed sql/platform/create_redemption.sql
	createRedemptionSQL string
)
