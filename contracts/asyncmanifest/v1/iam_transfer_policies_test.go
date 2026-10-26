package asyncmanifestv1

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestIAMSystemTransferPoliciesAreExact(t *testing.T) {
	export := IAMSystemRecordExportJobPolicy()
	if export.Kind != IAMSystemRecordExportJobKind || export.PayloadRevision != 1 ||
		export.OwnerDomain != "iam" || export.EffectClass != jobsv1.EffectExternal ||
		export.MaximumAttempts != 3 || export.Validate() != nil {
		t.Fatalf("IAM export policy = %#v", export)
	}
	importPolicy := IAMSystemRecordImportJobPolicy()
	if importPolicy.Kind != IAMSystemRecordImportJobKind ||
		importPolicy.PayloadRevision != 1 || importPolicy.OwnerDomain != "iam" ||
		importPolicy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		importPolicy.MaximumAttempts != 8 || importPolicy.Validate() != nil {
		t.Fatalf("IAM import policy = %#v", importPolicy)
	}
	reconcile := IAMSystemExportReconcileJobPolicy()
	if reconcile.Kind != IAMSystemExportReconcileJobKind ||
		reconcile.PayloadRevision != 1 || reconcile.OwnerDomain != "iam" ||
		reconcile.EffectClass != jobsv1.EffectExternal ||
		reconcile.MaximumAttempts != 8 || reconcile.Validate() != nil {
		t.Fatalf("IAM export reconciliation policy = %#v", reconcile)
	}
}
