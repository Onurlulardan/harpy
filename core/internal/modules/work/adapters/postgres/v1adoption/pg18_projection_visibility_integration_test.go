//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

func TestWorkPG18AdoptionProjectionVisibilityIsExactAndFailClosed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := startWorkPG18(t, ctx)
	prepareWorkSchemas(t, ctx, database.owner)
	schema := "work_adopted_a"
	targetID, sourceID := workUUID(91, 1), workUUID(91, 2)
	versionID := workUUID(91, 3)
	principal := "USER:" + workUUID(91, 4).String()
	record := projectionVisibilityRecord(t, sourceID.String(), versionID.String(), principal)
	seedProjectionVisibilityStage(t, ctx, database.owner, schema,
		workUUID(91, 5), workUUID(91, 6), targetID, sourceID, record)
	grantProjectionVisibilityFixture(t, ctx, database.owner, schema)

	connection, err := database.owner.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if _, err = connection.ExecContext(ctx, `SET SESSION AUTHORIZATION harpy_adopter;
SET search_path TO work_adopted_a,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(ctx,
		"SELECT work_v1_adoption_project_collaboration_v1('WORKFLOW_PROJECTION',$1,$2::jsonb)",
		targetID, record); err != nil {
		t.Fatalf("project workflow visibility: %v", err)
	}
	assertProjectionPrincipals(t, ctx, database.owner, schema, targetID, []string{principal})

	seedProjectionVisibilityStage(t, ctx, database.owner, schema,
		workUUID(91, 7), workUUID(91, 8), targetID, sourceID, record)
	if _, err = connection.ExecContext(ctx,
		"SELECT work_v1_adoption_project_collaboration_v1('WORKFLOW_PROJECTION',$1,$2::jsonb)",
		targetID, record); err == nil {
		t.Fatal("ambiguous active Work adoption visibility was accepted")
	}
	assertProjectionPrincipals(t, ctx, database.owner, schema, targetID, []string{principal})
}

func projectionVisibilityRecord(t *testing.T, sourceID, versionID, principal string) string {
	t.Helper()
	document, err := json.Marshal(map[string]any{
		"DueAt": "2026-08-03T09:00:00Z", "PermissionRevision": 7,
		"Revision": 3, "SourceID": sourceID, "Summary": "Approve order",
		"VersionID": versionID, "VisibleToPrincipals": []string{principal},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(document)
}

func seedProjectionVisibilityStage(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, operationID, runID, targetID, sourceID any, record string,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		if _, err := transaction.ExecContext(ctx, `
INSERT INTO work_adoption_runs(run_id,cutover_generation,manifest_sha256,transform_version,
 state,source_count,target_count,created_at,updated_at)
VALUES($1,9,pg_catalog.decode(pg_catalog.repeat('11',32),'hex'),'work_adoption_v1',
	'COPYING',1,1,transaction_timestamp(),transaction_timestamp())`, runID); err != nil {
			t.Fatal(err)
		}
		if _, err := transaction.ExecContext(ctx, `
INSERT INTO work_v1_adoption_stage(operation_id,run_id,source_kind,source_id,target_kind,
 target_id,source_sha256,target_sha256,target_record,removed,source_xid,commit_lsn,
 batch_ordinal,change_ordinal,updated_at)
VALUES($1,$2,'board_tasks',$3,'WORKFLOW_PROJECTION',$4,
 pg_catalog.decode(pg_catalog.repeat('22',32),'hex'),
 pg_catalog.sha256(pg_catalog.convert_to($5::jsonb::text,'UTF8')),$5::jsonb,false,
 '1'::xid8,'0/10'::pg_lsn,1,1,transaction_timestamp())`,
			operationID, runID, sourceID, targetID, record); err != nil {
			t.Fatal(err)
		}
	})
}

func grantProjectionVisibilityFixture(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx,
		"GRANT USAGE ON SCHEMA "+schema+" TO harpy_adopter; "+
			"GRANT EXECUTE ON FUNCTION "+schema+
			".work_v1_adoption_project_collaboration_v1(text,uuid,jsonb) TO harpy_adopter"); err != nil {
		t.Fatal(err)
	}
}

func assertProjectionPrincipals(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, targetID any, expected []string,
) {
	t.Helper()
	var document string
	if err := database.QueryRowContext(ctx,
		"SELECT pg_catalog.array_to_json(visible_to_principals)::text FROM "+
			schema+".work_inbox_items WHERE id=$1", targetID).Scan(&document); err != nil {
		t.Fatal(err)
	}
	var principals []string
	if err := json.Unmarshal([]byte(document), &principals); err != nil {
		t.Fatal(err)
	}
	if len(principals) != len(expected) || principals[0] != expected[0] {
		t.Fatalf("projection principals=%v want=%v", principals, expected)
	}
}
