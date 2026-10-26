package iam

import "testing"

func TestBillingAccessPermissionsAcceptOnlyDeclaredOperations(t *testing.T) {
	accepted := 0
	for permission, operations := range billingPermissionOperations {
		for _, operation := range operations {
			target, valid := resolveBillingPermission(permission, operation)
			if !valid || target.resource == "" || target.action == "" ||
				target.capability == "" {
				t.Fatalf("%s/%s target = %#v valid=%t",
					permission, operation, target, valid)
			}
			accepted++
		}
	}
	if accepted != 54 {
		t.Fatalf("accepted Billing permission-operation pairs = %d", accepted)
	}
	for _, rejected := range [][2]string{
		{"billing.read", "billing.subscription.change"},
		{"billing.manage", "billing.subscription.read"},
		{"platform.billing.catalog.read", "billing.report.subscription.list"},
		{"platform.billing.events.read", "billing.catalog.campaign.create"},
		{"platform.billing.coupons.manage", "billing.catalog.campaign.publish"},
		{"platform.billing.manage", "billing.contract.verify"},
	} {
		if _, valid := resolveBillingPermission(rejected[0], rejected[1]); valid {
			t.Fatalf("unexpected Billing access target: %v", rejected)
		}
	}
}

func TestBillingAccessPermissionsMapToExactIAMTargets(t *testing.T) {
	tests := []struct {
		permission, operation, resource, action, capability string
		global, command                                     bool
	}{
		{"billing.read", "billing.subscription.read", "billing", "read",
			"billing.subscription.read", false, false},
		{"billing.manage", "billing.subscription.change", "billing", "manage",
			"billing.subscription.manage", false, false},
		{"billing.usage.read", "billing.usage.read", "billing.usage", "read",
			"billing.usage.read", false, false},
		{"platform.billing.read", "billing.catalog.offering.list",
			"platform.billing", "read", "platform.billing.catalog.read", true, false},
		{"platform.billing.read", "billing.report.subscription.list",
			"platform.billing", "read", "platform.billing.subscriptions.read", true, false},
		{"platform.billing.manage", "billing.catalog.discount.publish",
			"platform.billing", "manage", "platform.billing.coupons.manage", true, true},
		{"platform.billing.audit.read", "billing.report.audit.list",
			"platform.billing.audit", "read", "platform.billing.events.read", true, false},
		{"platform.billing.revenue.read", "billing.report.revenue.read",
			"platform.billing.revenue", "read", "platform.billing.revenue.read", true, false},
	}
	for _, test := range tests {
		target, valid := resolveBillingPermission(
			test.permission, test.operation,
		)
		if !valid || target.resource != test.resource || target.action != test.action ||
			target.capability != test.capability || target.global != test.global ||
			target.command != test.command {
			t.Fatalf("permission %q operation %q target = %#v",
				test.permission, test.operation, target)
		}
	}
	if _, valid := resolveBillingPermission("platform.billing.*",
		"billing.catalog.offering.list"); valid {
		t.Fatal("wildcard Billing permission was accepted")
	}
	if _, valid := resolveBillingPermission("billing.read", "invalid"); valid {
		t.Fatal("invalid Billing operation was accepted")
	}
}
