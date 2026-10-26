package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMLifecycleDeliveryArtifactUUIDContractMigrationStaysPaired(t *testing.T) {
	const name = "20260725124800_allow_uuid_v8_lifecycle_delivery_artifacts"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("fresh and v1_adopted lifecycle artifact UUID repair differs")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect lifecycle artifact UUID repair: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"DROP CONSTRAINT iam_lifecycle_delivery_artifacts_ref_check",
		"ADD CONSTRAINT iam_lifecycle_delivery_artifacts_ref_check CHECK",
		"[1-8][0-9a-f]{3}",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("lifecycle artifact UUID repair is missing %q", marker)
		}
	}
}

func TestIAMLifecycleDeliveryWorkerStorageGrantStaysPaired(t *testing.T) {
	const name = "20260725124900_grant_iam_lifecycle_delivery_worker_storage"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("fresh and v1_adopted lifecycle worker storage grant differs")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect lifecycle worker storage grant: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"GRANT SELECT (",
		"ON iam_lifecycle_delivery_worker_receipts TO harpy_runtime",
		"ON iam_lifecycle_delivery_worker_claims TO harpy_runtime",
		"ON iam_lifecycle_delivery_artifacts TO harpy_runtime",
		"GRANT UPDATE (state, outcome_unknown_at)",
		"ON iam_one_time_tokens TO harpy_runtime",
		"ON iam_user_identifiers TO harpy_runtime",
		"GRANT DELETE ON iam_lifecycle_delivery_worker_claims",
		"GRANT DELETE ON iam_lifecycle_delivery_artifacts",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("lifecycle worker storage grant is missing %q", marker)
		}
	}
	if strings.Contains(source, "GRANT ALL") {
		t.Fatal("lifecycle worker storage grant exposes broad authority")
	}
}

func TestIAMLifecycleDeliveryDeleteRoutinesStayPaired(t *testing.T) {
	const name = "20260725125330_create_iam_lifecycle_delivery_delete_routines_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("fresh and v1_adopted lifecycle delete routines differ")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect lifecycle delete routines: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"REVOKE DELETE ON TABLE iam_lifecycle_delivery_worker_claims",
		"REVOKE DELETE ON TABLE iam_lifecycle_delivery_artifacts",
		"CREATE FUNCTION iam_lifecycle_delivery_release_claim_v1",
		"CREATE FUNCTION iam_lifecycle_delivery_delete_artifact_v1",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("lifecycle delete routine migration is missing %q", marker)
		}
	}
}
