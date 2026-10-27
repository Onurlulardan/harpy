package catalogquery

import _ "embed"

var (
	//go:embed sql/platform/list_offerings.sql
	listOfferingsSQL string
	//go:embed sql/platform/get_offering.sql
	getOfferingSQL string
	//go:embed sql/platform/get_active_campaign.sql
	getActiveCampaignSQL string
	//go:embed sql/platform/get_platform_offering.sql
	getPlatformOfferingSQL string
	//go:embed sql/platform/get_platform_discount.sql
	getPlatformDiscountSQL string
	//go:embed sql/platform/get_platform_campaign.sql
	getPlatformCampaignSQL string
	//go:embed sql/platform/resolve_legacy_identity.sql
	materializeLegacyIdentitySQL string
)
