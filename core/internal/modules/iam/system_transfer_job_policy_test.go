package iam

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestSystemTransferJobPoliciesAreIAMOwnedAndClosed(t *testing.T) {
	policies := SystemTransferJobPolicies()
	if len(policies) != 3 {
		t.Fatalf("policy count=%d", len(policies))
	}
	want := map[string]jobsv1.EffectClass{
		SystemExportJobKind:          jobsv1.EffectExternal,
		SystemImportJobKind:          jobsv1.EffectDatabaseTransactional,
		SystemExportReconcileJobKind: jobsv1.EffectExternal,
	}
	for _, policy := range policies {
		if policy.Validate() != nil || policy.OwnerDomain != "iam" ||
			policy.PayloadRevision != 1 || policy.EffectClass != want[policy.Kind] {
			t.Fatalf("policy=%+v", policy)
		}
	}
}
