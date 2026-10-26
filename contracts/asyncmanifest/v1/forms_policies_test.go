package asyncmanifestv1

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestFormsPublicationApplyResultPolicyIsOwnerBoundAndTransactional(t *testing.T) {
	policy := FormsPublicationApplyResultJobPolicy()
	if policy.Kind != FormsPublicationApplyResultJobKind || policy.PayloadRevision != 1 ||
		policy.OwnerDomain != "forms" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts || policy.Validate() != nil {
		t.Fatalf("Forms completion policy = %#v", policy)
	}
}

func TestFormsRecordExportPolicyMatchesExternalArtifactContract(t *testing.T) {
	policy := FormsRecordExportJobPolicy()
	if policy.Kind != FormsRecordExportJobKind || policy.PayloadRevision != 1 ||
		policy.OwnerDomain != "forms" || policy.EffectClass != jobsv1.EffectExternal ||
		policy.MaximumAttempts != 3 || policy.Validate() != nil {
		t.Fatalf("Forms record export policy = %#v", policy)
	}
}
