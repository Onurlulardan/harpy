package systemtransfer

import (
	"strings"
	"testing"
)

func TestImportAttemptClaimUsesPinnedV2Contract(t *testing.T) {
	for _, required := range []string{
		"iam_system_record_import_worker_claim_v2",
		"catalog_version_id",
		"catalog_sha256",
		"identity_manifest_sha256",
	} {
		if !strings.Contains(claimImportWorkerSQL, required) {
			t.Fatalf("import attempt claim lacks %q", required)
		}
	}
	if strings.Contains(claimImportWorkerSQL,
		"iam_system_record_import_worker_claim_v1(") {
		t.Fatal("import attempt claim still calls the V1 contract")
	}
}
