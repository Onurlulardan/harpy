package iam

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestSystemTransferProcessorPoliciesAreExact(t *testing.T) {
	if err := ValidateSystemExportJobPolicy(SystemExportJobPolicy()); err != nil {
		t.Fatalf("export policy: %v", err)
	}
	if err := ValidateSystemImportJobPolicy(SystemImportJobPolicy()); err != nil {
		t.Fatalf("import policy: %v", err)
	}
	if err := ValidateSystemExportReconcileJobPolicy(
		SystemExportReconcileJobPolicy(),
	); err != nil {
		t.Fatalf("export reconciliation policy: %v", err)
	}
	tests := []struct {
		name     string
		policy   jobsv1.KindPolicy
		validate func(jobsv1.KindPolicy) error
	}{
		{name: "export effect", policy: mustTransferPolicy(t, SystemExportJobKind,
			jobsv1.EffectDatabaseTransactional, 3), validate: ValidateSystemExportJobPolicy},
		{name: "import attempts", policy: mustTransferPolicy(t, SystemImportJobKind,
			jobsv1.EffectDatabaseTransactional, 7), validate: ValidateSystemImportJobPolicy},
		{name: "reconcile effect", policy: mustTransferPolicy(t,
			SystemExportReconcileJobKind, jobsv1.EffectDatabaseTransactional, 8),
			validate: ValidateSystemExportReconcileJobPolicy},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.validate(test.policy) != ErrSystemTransferJobPolicy {
				t.Fatalf("policy accepted: %+v", test.policy)
			}
		})
	}
}

func mustTransferPolicy(t *testing.T, kind string, effect jobsv1.EffectClass,
	attempts int) jobsv1.KindPolicy {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(kind, 1, "iam", effect, attempts)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}
