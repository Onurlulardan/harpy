package licenseauthority

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"

type Limit = billingapi.LicenseLimit

type EntitlementProjection = billingapi.LicenseEntitlementProjection

func ProjectEntitlements(snapshot VerifiedSnapshot) EntitlementProjection {
	limits := snapshot.Limits()
	return EntitlementProjection{
		Boolean: map[string]bool{},
		Limits: map[string]Limit{
			"iam.active_users":        projectLimit(limits.MaxUsers),
			"documents.storage_bytes": {Unlimited: true},
		},
		Channel: snapshot.Channel(),
	}
}

func projectLimit(value int64) Limit {
	if value == 0 {
		return Limit{Unlimited: true}
	}
	return Limit{Value: value}
}
