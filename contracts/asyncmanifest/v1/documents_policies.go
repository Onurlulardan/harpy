package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	DocumentsArchiveBuildJobKind    = "documents.archive.build"
	DocumentsArchiveExpireJobKind   = "documents.archive.expire"
	DocumentsBlobDeleteJobKind      = "documents.blob.delete"
	DocumentsItemPurgeJobKind       = "documents.item.purge"
	DocumentsUploadReconcileJobKind = "documents.upload.reconcile"
	DocumentsAdoptionCopyJobKind    = "documents.adoption.copy"
)

func DocumentsJobPolicies() []jobsv1.KindPolicy {
	definitions := []struct {
		kind   string
		effect jobsv1.EffectClass
	}{
		{DocumentsArchiveBuildJobKind, jobsv1.EffectExternal},
		{DocumentsArchiveExpireJobKind, jobsv1.EffectDatabaseTransactional},
		{DocumentsBlobDeleteJobKind, jobsv1.EffectExternal},
		{DocumentsItemPurgeJobKind, jobsv1.EffectDatabaseTransactional},
		{DocumentsUploadReconcileJobKind, jobsv1.EffectExternal},
		{DocumentsAdoptionCopyJobKind, jobsv1.EffectExternal},
	}
	result := make([]jobsv1.KindPolicy, 0, len(definitions))
	for _, definition := range definitions {
		policy, _ := jobsv1.NewKindPolicy(definition.kind, 1, "documents",
			definition.effect, jobsv1.DefaultMaximumAttempts)
		result = append(result, policy)
	}
	return result
}
