package iam

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	SystemExportJobKind          = "iam.system-resource.record-export"
	SystemImportJobKind          = "iam.system-resource.record-import"
	SystemExportReconcileJobKind = "iam.system-resource.export-reconcile"
)

func SystemExportJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(SystemExportJobKind, 1, "iam",
		jobsv1.EffectExternal, 3)
	return policy
}

func SystemImportJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(SystemImportJobKind, 1, "iam",
		jobsv1.EffectDatabaseTransactional, 8)
	return policy
}

func SystemExportReconcileJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(SystemExportReconcileJobKind, 1, "iam",
		jobsv1.EffectExternal, 8)
	return policy
}

func SystemTransferJobPolicies() []jobsv1.KindPolicy {
	return []jobsv1.KindPolicy{SystemExportJobPolicy(), SystemImportJobPolicy(),
		SystemExportReconcileJobPolicy()}
}
