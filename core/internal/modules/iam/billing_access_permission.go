package iam

import (
	"strings"
)

type billingPermissionTarget struct {
	resource   string
	action     string
	global     bool
	command    bool
	capability string
}

var billingPermissionTargets = map[string]billingPermissionTarget{
	"billing.read":                          {resource: "billing", action: "read", capability: "billing.subscription.read"},
	"billing.manage":                        {resource: "billing", action: "manage", capability: "billing.subscription.manage"},
	"billing.subscription.read":             {resource: "billing", action: "read", capability: "billing.subscription.read"},
	"billing.subscription.manage":           {resource: "billing", action: "manage", capability: "billing.subscription.manage"},
	"billing.usage.read":                    {resource: "billing.usage", action: "read", capability: "billing.usage.read"},
	"billing.entitlements.read":             {resource: "billing.entitlements", action: "read", capability: "billing.entitlements.read"},
	"platform.billing.catalog.read":         {resource: "platform.billing.catalog", action: "read", global: true, capability: "platform.billing.catalog.read"},
	"platform.billing.catalog.manage":       {resource: "platform.billing.catalog", action: "manage", global: true, command: true, capability: "platform.billing.catalog.manage"},
	"platform.billing.subscriptions.read":   {resource: "platform.billing.subscriptions", action: "read", global: true, capability: "platform.billing.subscriptions.read"},
	"platform.billing.subscriptions.manage": {resource: "platform.billing.subscriptions", action: "manage", global: true, command: true, capability: "platform.billing.subscriptions.manage"},
	"platform.billing.events.read":          {resource: "platform.billing.events", action: "read", global: true, capability: "platform.billing.events.read"},
	"platform.billing.revenue.read":         {resource: "platform.billing.revenue", action: "read", global: true, capability: "platform.billing.revenue.read"},
	"platform.billing.coupons.manage":       {resource: "platform.billing.coupons", action: "manage", global: true, command: true, capability: "platform.billing.coupons.manage"},
	"platform.billing.campaigns.manage":     {resource: "platform.billing.campaigns", action: "manage", global: true, command: true, capability: "platform.billing.campaigns.manage"},
}

var billingPermissionOperations = map[string][]string{
	"billing.read": {
		"billing.subscription.read", "billing.discount.validate",
	},
	"billing.manage": {
		"billing.subscription.change", "billing.subscription.cancel",
		"billing.subscription.resume", "billing.discount.redeem",
	},
	"billing.subscription.read": {"billing.subscription.read"},
	"billing.subscription.manage": {
		"billing.subscription.change", "billing.subscription.cancel",
		"billing.subscription.resume", "billing.discount.redeem",
	},
	"billing.usage.read":        {"billing.usage.read"},
	"billing.entitlements.read": {"billing.entitlements.read"},
	"platform.billing.catalog.read": {
		"billing.catalog.offering.list", "billing.catalog.offering.read",
		"billing.catalog.discount.list", "billing.catalog.discount.read",
		"billing.catalog.campaign.list", "billing.catalog.campaign.read",
	},
	"platform.billing.catalog.manage": {
		"billing.catalog.offering.create", "billing.catalog.offering.publish",
		"billing.catalog.offering.retire",
	},
	"platform.billing.subscriptions.read": {"billing.report.subscription.list"},
	"platform.billing.subscriptions.manage": {
		"billing.report.subscription.transition",
	},
	"platform.billing.events.read":  {"billing.report.audit.list"},
	"platform.billing.revenue.read": {"billing.report.revenue.read"},
	"platform.billing.coupons.manage": {
		"billing.catalog.discount.list", "billing.catalog.discount.read",
		"billing.catalog.discount.create", "billing.catalog.discount.publish",
		"billing.catalog.discount.retire",
	},
	"platform.billing.campaigns.manage": {
		"billing.catalog.campaign.list", "billing.catalog.campaign.read",
		"billing.catalog.campaign.create", "billing.catalog.campaign.publish",
		"billing.catalog.campaign.retire",
	},
	"platform.billing.read": {
		"billing.catalog.offering.list", "billing.catalog.offering.read",
		"billing.catalog.discount.list", "billing.catalog.discount.read",
		"billing.catalog.campaign.list", "billing.catalog.campaign.read",
		"billing.report.subscription.list",
	},
	"platform.billing.manage": {
		"billing.catalog.offering.create", "billing.catalog.offering.publish",
		"billing.catalog.offering.retire", "billing.catalog.discount.create",
		"billing.catalog.discount.publish", "billing.catalog.discount.retire",
		"billing.catalog.campaign.create", "billing.catalog.campaign.publish",
		"billing.catalog.campaign.retire", "billing.report.subscription.transition",
	},
	"platform.billing.audit.read": {"billing.report.audit.list"},
}

func resolveBillingPermission(
	permission, operation string,
) (billingPermissionTarget, bool) {
	target, found := billingPermissionTargets[permission]
	if found {
		return target, billingPermissionOperation(permission, operation)
	}
	target, found = legacyGlobalBillingTarget(permission, operation)
	return target, found && billingPermissionOperation(permission, operation)
}

func billingPermissionOperation(permission, operation string) bool {
	for _, allowed := range billingPermissionOperations[permission] {
		if operation == allowed {
			return true
		}
	}
	return false
}

func legacyGlobalBillingTarget(
	permission, operation string,
) (billingPermissionTarget, bool) {
	if permission == "platform.billing.audit.read" {
		target := billingPermissionTargets["platform.billing.events.read"]
		target.resource, target.action = "platform.billing.audit", "read"
		return target, true
	}
	if permission == "platform.billing.read" {
		key := "platform.billing.catalog.read"
		if strings.HasPrefix(operation, "billing.report.subscription.") {
			key = "platform.billing.subscriptions.read"
		}
		target := billingPermissionTargets[key]
		target.resource, target.action = "platform.billing", "read"
		return target, true
	}
	if permission != "platform.billing.manage" {
		return billingPermissionTarget{}, false
	}
	key := "platform.billing.catalog.manage"
	switch {
	case strings.HasPrefix(operation, "billing.report.subscription."):
		key = "platform.billing.subscriptions.manage"
	case strings.HasPrefix(operation, "billing.catalog.discount."):
		key = "platform.billing.coupons.manage"
	case strings.HasPrefix(operation, "billing.catalog.campaign."):
		key = "platform.billing.campaigns.manage"
	}
	target := billingPermissionTargets[key]
	target.resource, target.action = "platform.billing", "manage"
	return target, true
}
