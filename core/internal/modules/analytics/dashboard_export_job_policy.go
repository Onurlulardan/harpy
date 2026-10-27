package analytics

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	DashboardExportRenderJobKind = "analytics.dashboard.export.render"
	DashboardExportDeleteJobKind = "analytics.dashboard.export.delete"
	DashboardExportExpireJobKind = "analytics.dashboard.export.expire"
)

func DashboardExportRenderJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(DashboardExportRenderJobKind, 1, "analytics",
		jobsv1.EffectExternal, 3)
	return policy
}

func DashboardExportDeleteJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(DashboardExportDeleteJobKind, 1, "analytics",
		jobsv1.EffectExternal, 3)
	return policy
}

func DashboardExportExpireJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(DashboardExportExpireJobKind, 1, "analytics",
		jobsv1.EffectExternal, 8)
	return policy
}

// DashboardExportJobPolicies is the Analytics-owned contribution consumed by
// the single integration owner when the release job catalog is normalized.
func DashboardExportJobPolicies() []jobsv1.KindPolicy {
	return []jobsv1.KindPolicy{DashboardExportRenderJobPolicy(),
		DashboardExportDeleteJobPolicy(), DashboardExportExpireJobPolicy()}
}

func validDashboardExportJobPolicy(policy jobsv1.KindPolicy, kind string) bool {
	expected := DashboardExportRenderJobPolicy()
	if kind == DashboardExportDeleteJobKind {
		expected = DashboardExportDeleteJobPolicy()
	} else if kind == DashboardExportExpireJobKind {
		expected = DashboardExportExpireJobPolicy()
	}
	return policy == expected
}
