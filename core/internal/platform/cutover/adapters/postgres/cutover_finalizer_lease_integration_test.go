//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func verifyFinalizerLeaseDenials(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	connection *pgx.Conn, evidence finalizationEvidence,
) {
	t.Helper()
	now := time.Now().UTC()
	tests := []struct {
		name  string
		field string
		value any
	}{
		{"duration", "max_duration_seconds", 121},
		{"issued ordering", "issued_at", now.Add(time.Second).Format(time.RFC3339Nano)},
		{"procedure signature", "finalization_procedure_signature", "platform.wrong(uuid)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invalid := evidence
			invalid.leaseDocument = replaceJSONField(
				t, evidence.leaseDocument, test.field, test.value,
			)
			_, err := callFinalization(ctx, connection, invalid)
			assertFinalizationError(t, err, "22000", "authorization differs")
		})
	}
	invalidRole := evidence
	invalidRole.leaseDocument = replaceNestedJSONField(
		t, evidence.leaseDocument, "role_lease", "connection_limit", 2,
	)
	_, err := callFinalization(ctx, connection, invalidRole)
	assertFinalizationError(t, err, "22000", "authorization differs")
	if _, err := database.admin.Exec(ctx,
		"ALTER ROLE harpy_cutover_finalizer CONNECTION LIMIT 2"); err != nil {
		t.Fatal(err)
	}
	_, err = callFinalization(ctx, connection, evidence)
	assertFinalizationError(t, err, "55000", "finalizer role lease is not active")
	setFinalizerRoleLease(t, ctx, database.admin, evidence.expiresAt)
	if _, err := database.admin.Exec(ctx,
		"ALTER ROLE harpy_cutover_finalizer INHERIT"); err != nil {
		t.Fatal(err)
	}
	_, err = callFinalization(ctx, connection, evidence)
	assertFinalizationError(t, err, "55000", "finalizer role lease is not active")
	setFinalizerRoleLease(t, ctx, database.admin, evidence.expiresAt)
	if _, err := database.admin.Exec(ctx,
		"GRANT harpy_adopter TO harpy_cutover_finalizer"); err != nil {
		t.Fatal(err)
	}
	_, err = callFinalization(ctx, connection, evidence)
	assertFinalizationError(t, err, "55000", "finalizer role lease is not active")
	if _, err := database.admin.Exec(ctx,
		"REVOKE harpy_adopter FROM harpy_cutover_finalizer"); err != nil {
		t.Fatal(err)
	}
}

func replaceJSONField(
	t *testing.T, document json.RawMessage, field string, value any,
) json.RawMessage {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	decoded[field] = value
	return mustIntegrationJSON(t, decoded)
}

func replaceNestedJSONField(
	t *testing.T, document json.RawMessage, object, field string, value any,
) json.RawMessage {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	nested, ok := decoded[object].(map[string]any)
	if !ok {
		t.Fatalf("JSON field %s is not an object", object)
	}
	nested[field] = value
	return mustIntegrationJSON(t, decoded)
}
