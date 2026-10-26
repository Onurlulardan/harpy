package reports

const Owner = "billing"

const (
	SubscriptionResource = "billing.subscriptions"
	AuditResource        = "billing.audit_events"
	LegacyEventResource  = "billing.legacy_events"
	RevenueResource      = "billing.revenue_projection"
)

const (
	SubscriptionOperation           = "billing.report.subscription.list"
	SubscriptionTransitionOperation = "billing.report.subscription.transition"
	AuditOperation                  = "billing.report.audit.list"
	RevenueOperation                = "billing.report.revenue.read"
)
