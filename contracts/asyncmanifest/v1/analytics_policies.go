package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	AnalyticsQueryRevalidateJobKind       = "analytics.query.revalidate"
	AnalyticsCacheWarmJobKind             = "analytics.cache.warm"
	AnalyticsPublicationRepairJobKind     = "analytics.publication.repair"
	AnalyticsRetentionJobKind             = "analytics.retention.purge"
	AnalyticsDashboardExportRenderJobKind = "analytics.dashboard.export.render"
	AnalyticsDashboardExportDeleteJobKind = "analytics.dashboard.export.delete"
	AnalyticsDashboardExportExpireJobKind = "analytics.dashboard.export.expire"
)

func AnalyticsJobPolicies() []jobsv1.KindPolicy {
	definitions := []struct {
		kind     string
		effect   jobsv1.EffectClass
		attempts int
	}{
		{AnalyticsQueryRevalidateJobKind, jobsv1.EffectDatabaseTransactional,
			jobsv1.DefaultMaximumAttempts},
		{AnalyticsCacheWarmJobKind, jobsv1.EffectDatabaseTransactional,
			jobsv1.DefaultMaximumAttempts},
		{AnalyticsPublicationRepairJobKind, jobsv1.EffectDatabaseTransactional,
			jobsv1.DefaultMaximumAttempts},
		{AnalyticsRetentionJobKind, jobsv1.EffectDatabaseTransactional,
			jobsv1.DefaultMaximumAttempts},
		{AnalyticsDashboardExportRenderJobKind, jobsv1.EffectExternal, 3},
		{AnalyticsDashboardExportDeleteJobKind, jobsv1.EffectExternal, 3},
		{AnalyticsDashboardExportExpireJobKind, jobsv1.EffectExternal, 8},
	}
	policies := make([]jobsv1.KindPolicy, len(definitions))
	for index, definition := range definitions {
		policies[index], _ = jobsv1.NewKindPolicy(definition.kind, 1, "analytics",
			definition.effect, definition.attempts)
	}
	return policies
}
