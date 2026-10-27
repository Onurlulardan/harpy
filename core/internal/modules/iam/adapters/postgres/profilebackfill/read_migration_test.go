package profilebackfill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicationReadsAreActorBoundAndTableLocked(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		create := readProfileBackfillMigration(t, origin,
			"20260716015937_create_iam_profile_backfill_publication_reads_v1")
		grant := readProfileBackfillMigration(t, origin,
			"20260716015938_grant_iam_profile_backfill_publication_reads_v1")
		for _, fragment := range []string{
			"SECURITY DEFINER", "iam_profile_actor_assert_v1(",
			"LOCK TABLE iam_users IN SHARE MODE",
			"LOCK TABLE iam_organizations IN SHARE MODE",
			"LOCK TABLE iam_roles IN SHARE MODE",
			"p_limit NOT BETWEEN 1 AND 250",
			"source_schema_version_id uuid",
			"row_schema.version <= source_schema.version",
			"REVOKE ALL ON FUNCTION iam_profile_publication_receipt_read_v1(",
			"REVOKE ALL ON FUNCTION iam_profile_backfill_source_resources_plan_v1(",
		} {
			if !strings.Contains(create, fragment) {
				t.Fatalf("%s create migration missing %q", origin, fragment)
			}
		}
		if strings.Contains(create,
			"resource.profile_schema_version_id = p_source_schema_id") {
			t.Fatalf("%s source inventory excludes historical live rows", origin)
		}
		for _, fragment := range []string{
			"GRANT EXECUTE ON FUNCTION iam_profile_publication_receipt_read_v1(",
			"GRANT EXECUTE ON FUNCTION iam_profile_backfill_source_resources_plan_v1(",
			") TO harpy_runtime;",
		} {
			if !strings.Contains(grant, fragment) {
				t.Fatalf("%s grant migration missing %q", origin, fragment)
			}
		}
	}
}

func TestStaticPublicationReadsOnlyCallGuardedRoutines(t *testing.T) {
	for name, document := range map[string]string{
		"source": readSQL, "receipt": readReceiptSQL,
	} {
		for _, table := range []string{
			"iam_users", "iam_organizations", "iam_roles",
			"iam_authorization_operation_receipts",
		} {
			if strings.Contains(document, table) {
				t.Fatalf("%s static SQL reads raw table %s", name, table)
			}
		}
	}
	if !strings.Contains(readSQL, "iam_profile_backfill_source_resources_plan_v1(") ||
		!strings.Contains(readReceiptSQL, "iam_profile_publication_receipt_read_v1(") {
		t.Fatal("static SQL does not call guarded publication routines")
	}
}

func TestLineagePlanCreateUsesForwardCumulativeRoutines(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		draft := readProfileBackfillMigration(t, origin,
			"20260716015950_create_iam_profile_backfill_draft_create_v2")
		source := readProfileBackfillMigration(t, origin,
			"20260716015951_create_iam_profile_backfill_source_page_v2")
		page := readProfileBackfillMigration(t, origin,
			"20260716015952_create_iam_profile_backfill_page_create_v2")
		plan := readProfileBackfillMigration(t, origin,
			"20260716015953_create_iam_profile_backfill_plan_create_v2")
		grant := readProfileBackfillMigration(t, origin,
			"20260716015954_grant_iam_profile_backfill_plan_create_v2")
		for _, fragment := range []string{
			"fields must evolve cumulatively",
			"stored.state = 'retired' AND field.value ->> 'state' <> 'retired'",
		} {
			if !strings.Contains(draft, fragment) {
				t.Fatalf("%s cumulative draft missing %q", origin, fragment)
			}
		}
		for _, fragment := range []string{
			"resource.profile_schema_version_id",
			"row_schema.version <= source_schema.version",
			"v_schema_id::text",
		} {
			if !strings.Contains(source, fragment) {
				t.Fatalf("%s lineage evidence missing %q", origin, fragment)
			}
		}
		if !strings.Contains(page, "iam_profile_backfill_source_page_v2(") ||
			!strings.Contains(plan, "iam_profile_backfill_draft_create_v2(") ||
			!strings.Contains(plan, "iam_profile_backfill_page_create_v2(") ||
			!strings.Contains(grant, "GRANT EXECUTE ON FUNCTION iam_profile_backfill_plan_create_v2(") {
			t.Fatalf("%s forward plan chain is incomplete", origin)
		}
	}
	if !strings.Contains(createSQL, "iam_profile_backfill_plan_create_v2(") {
		t.Fatal("static create SQL does not call plan create v2")
	}
}

func readProfileBackfillMigration(t *testing.T, origin, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "migrations", "tenant", origin,
		name+"."+origin+".sql")
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(document)
}
