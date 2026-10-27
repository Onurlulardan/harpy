package sessioncapability

import platformiam "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"

type DeploymentMode string
type TenantPurpose string

const (
	ModeOnPremise DeploymentMode = "onpremise"
	ModeSaaS      DeploymentMode = "saas"

	PurposeOnPremisePrimary TenantPurpose = "onpremise_primary"
	PurposeSaaSControl      TenantPurpose = "saas_control"
	PurposeSaaSCustomer     TenantPurpose = "saas_customer"
)

func (mode DeploymentMode) Valid() bool {
	return mode == ModeOnPremise || mode == ModeSaaS
}

func (purpose TenantPurpose) ValidFor(mode DeploymentMode) bool {
	return mode == ModeOnPremise && purpose == PurposeOnPremisePrimary ||
		mode == ModeSaaS &&
			(purpose == PurposeSaaSControl || purpose == PurposeSaaSCustomer)
}

var onPremisePlatformCapabilities = stringSet(
	"platform.recovery.read", "platform.recovery.enable",
	"platform.recovery.rotate", "platform.recovery.disable",
	"platform.recovery.anchor.read", "platform.recovery.anchor.rotate",
	"platform.recovery_sets.read", "platform.recovery_sets.create",
	"platform.recovery_sets.manifest_download", "platform.recovery_sets.delete",
	"platform.recovery_sets.restore", "platform.restore_operations.plan_download",
	"platform.license.read", "platform.license.refresh",
)

var saaSPlatformCapabilities = stringSet(
	"platform.tenants.read", "platform.tenants.create", "platform.tenants.update",
	"platform.tenants.suspend", "platform.tenants.reactivate",
	"platform.tenants.offboard", "platform.tenants.purge",
	"platform.tenants.admin_setup.reissue",
	"platform.billing.catalog.read", "platform.billing.catalog.manage",
	"platform.billing.subscriptions.read", "platform.billing.subscriptions.manage",
	"platform.billing.events.read", "platform.billing.revenue.read",
	"platform.billing.coupons.manage", "platform.billing.campaigns.manage",
	"platform.analytics.usage.read", "platform.bug_reports.read",
	"platform.bug_reports.manage", "platform.contact_messages.read",
	"platform.contact_messages.manage", "platform.ai.analytics.read",
	"platform.presence.read",
)

func ModeAllowsPlatform(mode DeploymentMode, capability string) bool {
	if !mode.Valid() || !platformiam.CapabilityKnown(capability) {
		return false
	}
	_, onPremiseOnly := onPremisePlatformCapabilities[capability]
	_, saaSOnly := saaSPlatformCapabilities[capability]
	if onPremiseOnly {
		return mode == ModeOnPremise
	}
	if saaSOnly {
		return mode == ModeSaaS
	}
	return true
}

func stringSet(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
