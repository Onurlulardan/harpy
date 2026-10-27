package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const CatalogOperationResumeJobKind = "catalog.operation.resume"

func CatalogJobPolicies() []jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(CatalogOperationResumeJobKind, 1, "catalog",
		jobsv1.EffectDatabaseTransactional, jobsv1.DefaultMaximumAttempts)
	return []jobsv1.KindPolicy{policy}
}
