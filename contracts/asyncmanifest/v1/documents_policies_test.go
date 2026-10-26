package asyncmanifestv1

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestDocumentsJobPoliciesAreClosed(t *testing.T) {
	want := []struct {
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
	policies := DocumentsJobPolicies()
	if len(policies) != len(want) {
		t.Fatalf("policy count = %d, want %d", len(policies), len(want))
	}
	for index, expected := range want {
		policy := policies[index]
		if policy.Kind != expected.kind || policy.OwnerDomain != "documents" ||
			policy.EffectClass != expected.effect || policy.PayloadRevision != 1 ||
			policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts || policy.Validate() != nil {
			t.Fatalf("policy %d = %#v", index, policy)
		}
	}
}
