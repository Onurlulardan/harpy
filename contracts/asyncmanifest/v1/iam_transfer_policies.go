package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	IAMSystemRecordExportJobKind    = "iam.system-resource.record-export"
	IAMSystemRecordImportJobKind    = "iam.system-resource.record-import"
	IAMSystemExportReconcileJobKind = "iam.system-resource.export-reconcile"
)

func IAMSystemRecordExportJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		IAMSystemRecordExportJobKind, 1, "iam", jobsv1.EffectExternal, 3,
	)
	return policy
}

func IAMSystemRecordImportJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		IAMSystemRecordImportJobKind, 1, "iam",
		jobsv1.EffectDatabaseTransactional, 8,
	)
	return policy
}

func IAMSystemExportReconcileJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		IAMSystemExportReconcileJobKind, 1, "iam", jobsv1.EffectExternal, 8,
	)
	return policy
}

func IAMSystemTransferJobPolicies() []jobsv1.KindPolicy {
	return []jobsv1.KindPolicy{
		IAMSystemRecordExportJobPolicy(), IAMSystemRecordImportJobPolicy(),
		IAMSystemExportReconcileJobPolicy(),
	}
}
