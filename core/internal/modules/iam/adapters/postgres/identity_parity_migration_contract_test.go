package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMIdentityParityMigrationsArePairedAndBounded(t *testing.T) {
	tests := []struct {
		name    string
		markers []string
	}{
		{"20260801015100_create_iam_installation_derived_id_v1", []string{
			"iam_installation_derived_id_v1", "harpy:iam:installation:v1",
			"pg_catalog.set_byte",
		}},
		{"20260801015101_upgrade_iam_installation_user_profile_v1", []string{
			"iam_installation_user_profile_upgrade_v1", "user-profile-schema-v2",
			"'first_name'", "'last_name'", "'phone'", "state = 'retired'",
		}},
		{"20260801015102_create_iam_installation_initialize_v4", []string{
			"iam_installation_initialize_v4", "iam_installation_initialize_v3",
			"iam_installation_user_profile_upgrade_v1",
		}},
		{"20260801015110_create_iam_identity_user_reads_v3", []string{
			"iam_identity_user_read_v3", "iam_identity_user_list_v3",
			"email_verified boolean", "access_modes text[]", "access.state = 'active'",
		}},
		{"20260801015111_create_iam_identity_organization_reads_v2", []string{
			"iam_identity_organization_read_v2", "owner_ids uuid[]",
			"member.membership_kind = 'owner'",
		}},
		{"20260801015120_create_iam_identity_bulk_delete_receipts_v1", []string{
			"iam_identity_bulk_delete_receipts", "iam_identity_bulk_delete_reserve_v1",
			"deleted_count BETWEEN 1 AND 100", "FOR SHARE",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", test.name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", test.name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted migrations differ")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM identity parity migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("invalid IAM identity parity migration: %v", err)
			}
			source := string(fresh)
			for _, marker := range test.markers {
				if !strings.Contains(source, marker) {
					t.Fatalf("migration lacks marker %q", marker)
				}
			}
		})
	}
}
