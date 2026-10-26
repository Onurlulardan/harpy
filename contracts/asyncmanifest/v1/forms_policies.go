package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	FormsPublicationApplyResultJobKind = "forms.publication.apply-result"
	FormsSchemaEvolutionRuntimeJobKind = "forms.schema-evolution.runtime"
	FormsRecordImportJobKind           = "forms.record-import"
	FormsRecordExportJobKind           = "forms.record-export"
)

func FormsPublicationApplyResultJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		FormsPublicationApplyResultJobKind,
		1,
		"forms",
		jobsv1.EffectDatabaseTransactional,
		jobsv1.DefaultMaximumAttempts,
	)
	return policy
}

func FormsRecordImportJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		FormsRecordImportJobKind,
		1,
		"forms",
		jobsv1.EffectDatabaseTransactional,
		8,
	)
	return policy
}

func FormsRecordExportJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		FormsRecordExportJobKind,
		1,
		"forms",
		jobsv1.EffectExternal,
		3,
	)
	return policy
}

func FormsSchemaEvolutionRuntimeJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		FormsSchemaEvolutionRuntimeJobKind,
		1,
		"forms",
		jobsv1.EffectDatabaseTransactional,
		jobsv1.DefaultMaximumAttempts,
	)
	return policy
}

func FormsJobPolicies() []jobsv1.KindPolicy {
	return []jobsv1.KindPolicy{
		FormsPublicationApplyResultJobPolicy(),
		FormsSchemaEvolutionRuntimeJobPolicy(),
		FormsRecordImportJobPolicy(),
		FormsRecordExportJobPolicy(),
	}
}
