package platformiam

var CapabilityCatalog = []string{
	"platform.instance.read",
	"platform.operators.read", "platform.operators.grant", "platform.operators.revoke",
	"platform.owner.transfer",
	"platform.tenants.read", "platform.tenants.create", "platform.tenants.update",
	"platform.tenants.suspend", "platform.tenants.reactivate",
	"platform.tenants.offboard", "platform.tenants.purge",
	"platform.tenants.admin_setup.reissue",
	"platform.database.stats.read", "platform.database.schema.read",
	"platform.database.diagnostics.read",
	"platform.recovery.read", "platform.recovery.enable",
	"platform.recovery.rotate", "platform.recovery.disable",
	"platform.recovery.anchor.read", "platform.recovery.anchor.rotate",
	"platform.recovery_sets.read", "platform.recovery_sets.create",
	"platform.recovery_sets.manifest_download", "platform.recovery_sets.delete",
	"platform.recovery_sets.restore", "platform.restore_operations.plan_download",
	"platform.billing.catalog.read", "platform.billing.catalog.manage",
	"platform.billing.subscriptions.read", "platform.billing.subscriptions.manage",
	"platform.billing.events.read", "platform.billing.revenue.read",
	"platform.billing.coupons.manage", "platform.billing.campaigns.manage",
	"platform.analytics.usage.read", "platform.bug_reports.read",
	"platform.bug_reports.manage", "platform.contact_messages.read",
	"platform.contact_messages.manage", "platform.operations.health.read",
	"platform.operations.alerts.read", "platform.operations.alerts.manage",
	"platform.operations.traces.read",
	"platform.operations.permission_cache.flush",
	"platform.license.read", "platform.license.refresh",
	"platform.ai.system_knowledge.read", "platform.ai.system_knowledge.reingest",
	"platform.ai.analytics.read", "platform.presence.read",
}

var capabilities = func() map[string]struct{} {
	result := make(map[string]struct{}, len(CapabilityCatalog))
	for _, capability := range CapabilityCatalog {
		result[capability] = struct{}{}
	}
	return result
}()

func CapabilityKnown(capability string) bool {
	_, known := capabilities[capability]
	return known
}
