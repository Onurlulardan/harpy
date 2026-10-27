package billing

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestBillingSnapshotReadMigrationsArePairedAndLockFree(t *testing.T) {
	cases := []struct {
		name    string
		markers []string
	}{
		{
			name: "20260728110005_create_billing_snapshot_authority_and_price_reads_v1",
			markers: []string{
				"CREATE FUNCTION platform.billing_resolve_current_global_authority_v1",
				"CREATE FUNCTION platform.billing_effective_subscription_price_snapshot_read_v1",
				"billing_resolve_current_global_authority_v1(text)",
				"billing_effective_subscription_price_snapshot_read_v1(",
			},
		},
		{
			name: "20260728110006_create_billing_catalog_snapshot_reads_v1",
			markers: []string{
				"CREATE VIEW platform.billing_catalog_offering_snapshot_read_v1",
				"CREATE VIEW platform.billing_catalog_campaign_snapshot_read_v1",
				"GRANT SELECT ON platform.billing_catalog_offering_snapshot_read_v1",
				"GRANT SELECT ON platform.billing_catalog_campaign_snapshot_read_v1",
			},
		},
		{
			name: "20260728110007_create_billing_subscription_snapshot_read_v1",
			markers: []string{
				"CREATE FUNCTION platform.billing_subscription_snapshot_read_v1",
				"billing_subscription_snapshot_read_v1(uuid,uuid)",
				"TO harpy_runtime",
			},
		},
		{
			name: "20260728110008_create_billing_usage_snapshot_reads_v1",
			markers: []string{
				"CREATE FUNCTION platform.billing_usage_context_snapshot_read_v1",
				"CREATE FUNCTION platform.billing_usage_snapshot_read_v1",
				"billing_usage_context_snapshot_read_v1(uuid,uuid)",
				"billing_usage_snapshot_read_v1(uuid,uuid,uuid,uuid,uuid)",
			},
		},
		{
			name: "20260728110009_create_billing_discount_snapshot_reads_v1",
			markers: []string{
				"CREATE FUNCTION platform.billing_discount_eligibility_snapshot_read_v1",
				"CREATE FUNCTION platform.billing_discount_offerings_snapshot_read_v1",
				"billing_discount_eligibility_snapshot_read_v1(",
				"billing_discount_offerings_snapshot_read_v1(uuid,uuid)",
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fresh := readBillingSnapshotMigration(
				t, "fresh", testCase.name+".fresh.sql",
			)
			if len(fresh) == 0 {
				t.Fatal("billing snapshot-read migration is empty")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("billing snapshot-read migration is invalid: %v", err)
			}
			source := string(fresh)
			for _, marker := range testCase.markers {
				if !strings.Contains(source, marker) {
					t.Errorf("billing snapshot-read migration lacks %q", marker)
				}
			}
			for _, forbidden := range []string{
				"FOR SHARE", "FOR UPDATE", "-- +goose Down",
			} {
				if strings.Contains(source, forbidden) {
					t.Errorf("billing snapshot-read migration contains %q", forbidden)
				}
			}
		})
	}
}
