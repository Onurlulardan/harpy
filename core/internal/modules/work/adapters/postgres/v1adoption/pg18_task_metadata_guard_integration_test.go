//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
)

func assertProtectedTaskMetadataDocumentGuard(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		_, err := transaction.ExecContext(ctx, `
CREATE FUNCTION work_test_task_document_valid_v1(p_record jsonb)
RETURNS boolean LANGUAGE sql SECURITY DEFINER
SET search_path FROM CURRENT
AS 'SELECT work_v1_adoption_document_valid_v1(''TASK'',p_record)';
REVOKE ALL ON FUNCTION work_test_task_document_valid_v1(jsonb) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION work_test_task_document_valid_v1(jsonb) TO harpy_adopter;`)
		if err != nil {
			t.Fatal(err)
		}
	})
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx,
		"SET LOCAL SESSION AUTHORIZATION harpy_adopter"); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)", schema+",pg_catalog"); err != nil {
		t.Fatal(err)
	}
	valid := protectedTaskGuardDocument()
	assertTaskGuardResult(t, ctx, transaction, valid, true, "valid")
	invalidRule := protectedTaskGuardDocument()
	invalidRule["RecurrenceRule"] = 42
	assertTaskGuardResult(t, ctx, transaction, invalidRule, false, "numeric recurrence rule")
	invalidEnd := protectedTaskGuardDocument()
	invalidEnd["RecurrenceEnd"] = map[string]any{"at": "2026-08-21T09:00:00Z"}
	assertTaskGuardResult(t, ctx, transaction, invalidEnd, false, "object recurrence end")
	oversizeRule := protectedTaskGuardDocument()
	oversizeRule["RecurrenceRule"] = strings.Repeat("ş", 51)
	assertTaskGuardResult(t, ctx, transaction, oversizeRule, false, "51 character rule")
	for name, filter := range map[string]any{"null filter": nil, "array filter": []any{},
		"scalar filter": "legacy"} {
		document := protectedTaskGuardDocument()
		document["FormFilter"] = filter
		assertTaskGuardResult(t, ctx, transaction, document, true, name)
	}
}

func assertTaskGuardResult(t *testing.T, ctx context.Context, transaction *sql.Tx,
	document map[string]any, expected bool, name string,
) {
	t.Helper()
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var actual bool
	if err = transaction.QueryRowContext(ctx,
		"SELECT work_test_task_document_valid_v1($1::jsonb)", body).Scan(&actual); err != nil {
		t.Fatalf("%s guard query: %v", name, err)
	}
	if actual != expected {
		t.Fatalf("%s guard result=%t want=%t", name, actual, expected)
	}
}

func protectedTaskGuardDocument() map[string]any {
	return map[string]any{
		"BoardID": "41000000-0000-4000-8000-000000000302", "CancelledAt": nil,
		"ColumnID": "41000000-0000-4000-8000-000000000304", "CompletedAt": nil,
		"CreatedAt": "2026-07-21T09:00:00Z",
		"CreatorID": "41000000-0000-4000-8000-000000000303", "DeletedAt": nil,
		"Description": "Legacy recurrence", "DueAt": "2026-07-22T09:00:00Z",
		"FormFilter": map[string]any{"status": "open"}, "IsRecurring": true,
		"Kind": "CUSTOM", "Position": 1024, "Priority": "HIGH",
		"RecurrenceEnd": "2026-08-21T09:00:00Z", "RecurrenceRule": "FREQ=WEEKLY",
		"StartsAt": "2026-07-21T09:00:00Z", "Status": "OPEN",
		"Title": "Protected task", "UpdatedAt": "2026-07-21T09:00:00Z",
	}
}
