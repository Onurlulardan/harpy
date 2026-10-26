package httpapi

import "context"

type PlatformCatalog interface {
	ListOfferings(context.Context, Query) (QueryPage, error)
	CreateOffering(context.Context, CreateOfferingCommand) (OfferingView, error)
	GetOffering(context.Context, Access, string) (OfferingView, error)
	CreateOfferingVersion(context.Context, CreateOfferingVersionCommand) (OfferingView, error)
	RetireOffering(context.Context, RetireOfferingCommand) (OfferingView, error)
	ListDiscounts(context.Context, Query) (QueryPage, error)
	CreateDiscount(context.Context, CreateDiscountCommand) (DiscountView, error)
	GetDiscount(context.Context, Access, string) (DiscountView, error)
	CreateDiscountVersion(context.Context, CreateDiscountVersionCommand) (DiscountView, error)
	RetireDiscount(context.Context, RetireDiscountCommand) (DiscountView, error)
	ListCampaigns(context.Context, Query) (QueryPage, error)
	CreateCampaign(context.Context, CreateCampaignCommand) (CampaignView, error)
	GetCampaign(context.Context, Access, string) (CampaignView, error)
	CreateCampaignVersion(context.Context, CreateCampaignVersionCommand) (CampaignView, error)
	RetireCampaign(context.Context, RetireCampaignCommand) (CampaignView, error)
}

type PlatformReports interface {
	ListSubscriptions(context.Context, Query) (Page[SubscriptionView], error)
	TransitionSubscription(context.Context, TransitionSubscriptionCommand) (SubscriptionView, error)
	ListAuditEvents(context.Context, Query) (Page[AuditEventView], error)
	ListLegacyEvents(context.Context, Query) (Page[LegacyEventView], error)
	GetRevenue(context.Context, Query) (RevenueResult, error)
}
