package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMTenantAuthorityAssertionIsReadOnlyTransactionSafe(t *testing.T) {
	const name = "20260725125000_remove_iam_tenant_authority_read_lock"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	adopted := readIAMPlatformAuthMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("fresh and v1_adopted tenant authority migration differs")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("tenant authority migration is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE OR REPLACE FUNCTION platform.iam_assert_current_tenant_authority_v1",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"FROM platform.cutover_groups", "FROM platform.domain_cutovers",
		"FROM platform.cutover_group_members",
		"REVOKE ALL ON FUNCTION platform.iam_assert_current_tenant_authority_v1",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("tenant authority migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{"FOR SHARE", "FOR UPDATE", "INSERT ", "UPDATE ", "DELETE "} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("tenant authority assertion is not read-only safe: %q", forbidden)
		}
	}
}
