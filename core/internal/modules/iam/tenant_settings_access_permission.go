package iam

type tenantSettingsAccessTarget struct{ resource, action string }

func resolveTenantSettingsAccess(permission, operation string) (
	tenantSettingsAccessTarget, bool,
) {
	target := tenantSettingsAccessTarget{resource: "iam.organizations"}
	switch {
	case permission == "tenancy.settings.read" && operation == "tenancy.settings.get":
		target.action = "read"
		return target, true
	case permission == "tenancy.settings.manage" &&
		(operation == "tenancy.settings.update" ||
			operation == "tenancy.settings.test_email"):
		target.action = "update"
		return target, true
	default:
		return tenantSettingsAccessTarget{}, false
	}
}
