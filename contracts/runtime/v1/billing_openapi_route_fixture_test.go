package runtimev1

func canonicalBillingOperations() map[string]billingOperation {
	return map[string]billingOperation{
		"GET /api/billing/v2/offerings":         {Method: "GET"},
		"GET /api/billing/v2/promotions/active": {Method: "GET"},
		"GET /api/billing/v2/subscriptions/current": {
			Method: "GET", Permission: "billing.read",
		},
		"GET /api/billing/v2/usage": {
			Method: "GET", Permission: "billing.usage.read",
		},
		"POST /api/billing/v2/subscriptions/current/changes": {
			Method: "POST", Permission: "billing.manage",
		},
		"POST /api/billing/v2/subscriptions/current/cancellations": {
			Method: "POST", Permission: "billing.manage",
		},
		"POST /api/billing/v2/subscriptions/current/cancellations/resume": {
			Method: "POST", Permission: "billing.manage",
		},
		"POST /api/billing/v2/subscriptions/current/discount-redemptions": {
			Method: "POST", Permission: "billing.manage",
		},
		"POST /api/billing/v2/discounts/validate": {
			Method: "POST", Permission: "billing.read",
		},
		"GET /api/platform/billing/offerings": {
			Method: "GET", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/offerings": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"POST /api/platform/billing/offerings/query": {
			Method: "POST", Permission: "platform.billing.read",
		},
		"GET /api/platform/billing/offerings/{offering_id}": {
			Method: "GET", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/offerings/{offering_id}/versions": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"POST /api/platform/billing/offerings/{offering_id}/retire": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"GET /api/platform/billing/subscriptions": {
			Method: "GET", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/subscriptions/query": {
			Method: "POST", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/subscriptions/{subscription_id}/transitions": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"GET /api/platform/billing/events": {
			Method: "GET", Permission: "platform.billing.audit.read",
		},
		"POST /api/platform/billing/events/query": {
			Method: "POST", Permission: "platform.billing.audit.read",
		},
		"POST /api/platform/billing/revenue/query": {
			Method: "POST", Permission: "platform.billing.revenue.read",
		},
		"GET /api/platform/billing/discounts": {
			Method: "GET", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/discounts": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"POST /api/platform/billing/discounts/query": {
			Method: "POST", Permission: "platform.billing.read",
		},
		"GET /api/platform/billing/discounts/{discount_id}": {
			Method: "GET", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/discounts/{discount_id}/versions": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"POST /api/platform/billing/discounts/{discount_id}/retire": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"GET /api/platform/billing/campaigns": {
			Method: "GET", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/campaigns": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"POST /api/platform/billing/campaigns/query": {
			Method: "POST", Permission: "platform.billing.read",
		},
		"GET /api/platform/billing/campaigns/{campaign_id}": {
			Method: "GET", Permission: "platform.billing.read",
		},
		"POST /api/platform/billing/campaigns/{campaign_id}/versions": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"POST /api/platform/billing/campaigns/{campaign_id}/retire": {
			Method: "POST", Permission: "platform.billing.manage",
		},
		"POST /webhooks/billing/{provider}": {Method: "POST"},
	}
}
